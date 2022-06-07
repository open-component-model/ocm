// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package accessio

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/marstr/guid"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/opencontainers/go-digest"
)

type Allocatable interface {
	Ref() error
	Unref() error
}

type BlobSource interface {
	Allocatable
	GetBlobData(digest digest.Digest) (DataAccess, error)
	GetBlob(digest digest.Digest) (int64, DataAccess, error)
}

type BlobSink interface {
	Allocatable
	AddBlob(blob BlobAccess) (int64, digest.Digest, error)
}

type BlobCache interface {
	BlobSource
	BlobSink
	AddData(data DataAccess) (int64, digest.Digest, error)
}

type blobCache struct {
	lock     sync.RWMutex
	cache    vfs.FileSystem
	refcount int
}

func NewDefaultBlobCache(fss ...vfs.FileSystem) (BlobCache, error) {
	var err error
	fs := DefaultedFileSystem(nil, fss...)
	if fs == nil {
		fs, err = osfs.NewTempFileSystem()
		if err != nil {
			return nil, err
		}
	}
	return &blobCache{
		cache:    fs,
		refcount: 1,
	}, nil
}

func NewStaticBlobCache(path string, fss ...vfs.FileSystem) (BlobCache, error) {
	fs := FileSystem(fss...)
	err := fs.MkdirAll(path, 0700)
	if err != nil {
		return nil, err
	}
	fs, err = projectionfs.New(fs, path)
	if err != nil {
		return nil, err
	}
	return &blobCache{
		cache:    fs,
		refcount: 1,
	}, nil
}

func (c *blobCache) Ref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cache == nil {
		return ErrClosed
	}
	c.refcount++
	return nil
}

func (c *blobCache) Unref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cache == nil {
		return ErrClosed
	}
	c.refcount--
	var err error
	if c.refcount <= 0 {
		err = vfs.Cleanup(c.cache)
		c.cache = nil
	}
	return err
}

func (c *blobCache) GetBlobData(digest digest.Digest) (DataAccess, error) {
	_, acc, err := c.GetBlob(digest)
	return acc, err
}

func (c *blobCache) GetBlob(digest digest.Digest) (int64, DataAccess, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.cache == nil {
		return -1, nil, ErrClosed
	}

	path := common.DigestToFileName(digest)
	fi, err := c.cache.Stat(path)
	if err == nil {
		return fi.Size(), DataAccessForFile(c.cache, path), nil
	}
	if os.IsNotExist(err) {
		return -1, nil, ErrBlobNotFound(digest)
	}
	return BLOB_UNKNOWN_SIZE, nil, err
}

func (c *blobCache) AddBlob(blob BlobAccess) (int64, digest.Digest, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cache == nil {
		return BLOB_UNKNOWN_SIZE, "", ErrClosed
	}

	var digester *DigestReader
	var path string

	if blob.DigestKnown() {
		path = common.DigestToFileName(blob.Digest())
		if ok, err := vfs.Exists(c.cache, path); ok || err != nil {
			return blob.Size(), blob.Digest(), err
		}
	} else {
		path = "TMP" + guid.NewGUID().String()
	}

	br, err := blob.Reader()
	if err != nil {
		return BLOB_UNKNOWN_SIZE, "", errors.Wrapf(err, "cannot get blob content")
	}
	defer br.Close()

	reader := io.Reader(br)
	if !blob.DigestKnown() {
		digester = NewDefaultDigestReader(reader)
		reader = digester
	}

	writer, err := c.cache.Create(path)
	if err != nil {
		return BLOB_UNKNOWN_SIZE, "", errors.Wrapf(err, "cannot create blob file in cache")
	}
	defer writer.Close()
	size, err := io.Copy(writer, reader)
	if err != nil {
		c.cache.Remove(path)
		return BLOB_UNKNOWN_SIZE, "", err
	}
	if digester != nil {
		target := common.DigestToFileName(digester.Digest())
		if ok, err := vfs.Exists(c.cache, target); err != nil || !ok {
			err = c.cache.Rename(path, target)
		}
		c.cache.Remove(path)
		return size, digester.Digest(), err
	}
	return size, blob.Digest(), nil
}

func (c *blobCache) AddData(data DataAccess) (int64, digest.Digest, error) {
	return c.AddBlob(BlobAccessForDataAccess(BLOB_UNKNOWN_DIGEST, BLOB_UNKNOWN_SIZE, "", data))
}

////////////////////////////////////////////////////////////////////////////////

type cascadedCache struct {
	lock     sync.RWMutex
	parent   BlobSource
	source   BlobSource
	sink     BlobSink
	refcount int
}

var _ BlobCache = (*cascadedCache)(nil)

func NewCascadedBlobCache(parent BlobCache) (BlobCache, error) {
	if parent != nil {
		err := parent.Ref()
		if err != nil {
			return nil, err
		}
	}
	return &cascadedCache{
		parent:   parent,
		refcount: 1,
	}, nil
}

func NewCascadedBlobCacheForSource(parent BlobSource, src BlobSource) (BlobCache, error) {
	if parent != nil {
		err := parent.Ref()
		if err != nil {
			return nil, err
		}
	}
	if src != nil {
		err := src.Ref()
		if err != nil {
			return nil, err
		}
	}
	return &cascadedCache{
		parent:   parent,
		source:   src,
		refcount: 1,
	}, nil
}

func NewCascadedBlobCacheForCache(parent BlobSource, src BlobCache) (BlobCache, error) {
	if parent != nil {
		err := parent.Ref()
		if err != nil {
			return nil, err
		}
	}
	if src != nil {
		err := src.Ref()
		if err != nil {
			return nil, err
		}
	}
	return &cascadedCache{
		parent: parent,
		source: src,
		sink:   src,
	}, nil
}

func (c *cascadedCache) Ref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcount == 0 {
		return ErrClosed
	}
	c.refcount++
	return nil
}

func (c *cascadedCache) Unref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcount == 0 {
		return ErrClosed
	}
	c.refcount--
	list := errors.ErrListf("closing cascaded blob cache")
	if c.refcount <= 0 {
		if c.source != nil {
			list.Add(c.source.Unref())
		}
		if c.parent != nil {
			list.Add(c.parent.Unref())
		}
	}
	return list.Result()
}

func (c *cascadedCache) GetBlobData(digest digest.Digest) (DataAccess, error) {
	_, acc, err := c.GetBlob(digest)
	return acc, err
}

func (c *cascadedCache) GetBlob(digest digest.Digest) (int64, DataAccess, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.refcount == 0 {
		return -1, nil, ErrClosed
	}

	if c.source != nil {
		size, acc, err := c.source.GetBlob(digest)
		if err == nil {
			return size, acc, err
		}
		if !IsErrBlobNotFound(err) {
			return -1, nil, err
		}
	}
	if c.parent != nil {
		return c.parent.GetBlob(digest)
	}
	return BLOB_UNKNOWN_SIZE, nil, ErrBlobNotFound(digest)
}

func (c *cascadedCache) AddData(data DataAccess) (int64, digest.Digest, error) {
	return c.AddBlob(BlobAccessForDataAccess(BLOB_UNKNOWN_DIGEST, BLOB_UNKNOWN_SIZE, "", data))
}

func (c *cascadedCache) AddBlob(blob BlobAccess) (int64, digest.Digest, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcount == 0 {
		return -1, "", ErrClosed
	}

	if c.source == nil {
		cache, err := NewDefaultBlobCache()
		if err != nil {
			return -1, "", err
		}
		c.source = cache
		c.sink = cache
	}
	if c.sink != nil {
		return c.sink.AddBlob(blob)
	}
	if c.parent != nil {
		if sink, ok := c.parent.(BlobSink); ok {
			return sink.AddBlob(blob)
		}
	}
	return -1, "", ErrReadOnly
}

////////////////////////////////////////////////////////////////////////////////

type cached struct {
	lock     sync.RWMutex
	source   BlobSource
	sink     BlobSink
	cache    BlobCache
	refcount int
}

var _ BlobCache = (*cached)(nil)

func CachedAccess(src BlobSource, dst BlobSink, cache BlobCache) (BlobCache, error) {
	var err error
	if cache == nil {
		cache, err = NewDefaultBlobCache()
		if err != nil {
			return nil, err
		}
	} else {
		err = cache.Ref()
		if err != nil {
			return nil, err
		}
	}
	if src != nil {
		err = src.Ref()
		if err != nil {
			return nil, err
		}
	}
	if dst != nil {
		err = dst.Ref()
		if err != nil {
			return nil, err
		}
	}
	return &cached{source: src, sink: dst, cache: cache, refcount: 1}, nil
}

type cachedAccess struct {
	lock   sync.Mutex
	cache  *cached
	access DataAccess
	digest digest.Digest
	size   int64
	orig   DataAccess
}

func newCachedAccess(cache *cached, blob DataAccess, size int64, digest digest.Digest) DataAccess {
	return &cachedAccess{
		cache:  cache,
		size:   size,
		digest: digest,
		orig:   blob,
	}
}

func (c *cachedAccess) Get() ([]byte, error) {
	var err error

	c.lock.Lock()
	defer c.lock.Unlock()
	if c.access == nil && c.digest != "" {
		c.size, c.access, _ = c.cache.cache.GetBlob(c.digest)
	}
	if c.access == nil {
		c.cache.lock.Lock()
		defer c.cache.lock.Unlock()

		if c.digest != "" {
			c.size, c.access, err = c.cache.cache.GetBlob(c.digest)
			if err != nil && !IsErrBlobNotFound(err) {
				return nil, err
			}
		}
		if c.access == nil {
			data, err := c.orig.Get()
			if err != nil {
				return nil, err
			}
			c.size, c.digest, err = c.cache.cache.AddData(DataAccessForBytes(data))
			if err == nil {
				c.orig.Close()
				c.orig = nil
			}
			return data, err
		}
	}
	return c.access.Get()
}

func (c cachedAccess) Reader() (io.ReadCloser, error) {
	var err error

	c.lock.Lock()
	defer c.lock.Unlock()
	if c.access == nil && c.digest != "" {
		c.size, c.access, _ = c.cache.cache.GetBlob(c.digest)
	}
	if c.access == nil {
		c.cache.lock.Lock()
		defer c.cache.lock.Unlock()

		if c.digest != "" {
			c.size, c.access, err = c.cache.cache.GetBlob(c.digest)
			if err != nil && !IsErrBlobNotFound(err) {
				return nil, err
			}
		}
		if c.access == nil {
			c.size, c.digest, err = c.cache.cache.AddData(c.orig)
			if err == nil {
				_, c.access, err = c.cache.cache.GetBlob(c.digest)
			}
			if err != nil {
				return nil, err
			}
			c.orig.Close()
			c.orig = nil
		}
	}
	return c.access.Reader()
}

func (c *cachedAccess) Close() error {
	return nil
}

func (c *cachedAccess) Size() int64 {
	return c.size
}

var _ DataAccess = (*cachedAccess)(nil)

func (c *cached) Ref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcount == 0 {
		return ErrClosed
	}
	c.refcount++
	return nil
}

func (c *cached) Unref() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.refcount == 0 {
		return ErrClosed
	}
	c.refcount--
	list := errors.ErrListf("closing cached blob store")
	if c.refcount <= 0 {
		if c.sink != nil {
			list.Add(c.sink.Unref())
		}
		if c.source != nil {
			list.Add(c.source.Unref())
		}
		c.cache.Unref()
	}
	return list.Result()
}

func (a *cached) GetBlobData(digest digest.Digest) (DataAccess, error) {
	acc, err := a.cache.GetBlobData(digest)
	if err != nil {
		if !IsErrBlobNotFound(err) {
			return nil, err
		}
		var size int64
		size, acc, err = a.source.GetBlob(digest)
		if err == nil {
			acc = newCachedAccess(a, acc, size, digest)
		}
	}
	return acc, err
}

func (a *cached) GetBlob(digest digest.Digest) (int64, DataAccess, error) {
	size, acc, err := a.cache.GetBlob(digest)
	if err != nil {
		if !IsErrBlobNotFound(err) {
			return BLOB_UNKNOWN_SIZE, nil, err
		}
		size, acc, err = a.source.GetBlob(digest)
		if err != nil {
			acc = newCachedAccess(a, acc, size, digest)
		}
	}
	return size, acc, err
}

func (a *cached) AddBlob(blob BlobAccess) (int64, digest.Digest, error) {
	if a.sink == nil {
		return BLOB_UNKNOWN_SIZE, BLOB_UNKNOWN_DIGEST, fmt.Errorf("no blob sink")
	}
	size, digest, err := a.cache.AddBlob(blob)
	if err != nil {
		return BLOB_UNKNOWN_SIZE, BLOB_UNKNOWN_DIGEST, err
	}
	acc, err := a.cache.GetBlobData(digest)
	if err != nil {
		return BLOB_UNKNOWN_SIZE, BLOB_UNKNOWN_DIGEST, err
	}
	size, digest, err = a.sink.AddBlob(BlobAccessForDataAccess(digest, size, blob.MimeType(), acc))
	if err != nil {
		return BLOB_UNKNOWN_SIZE, BLOB_UNKNOWN_DIGEST, err
	}
	return size, digest, err
}

func (c *cached) AddData(data DataAccess) (int64, digest.Digest, error) {
	return c.AddBlob(BlobAccessForDataAccess(BLOB_UNKNOWN_DIGEST, BLOB_UNKNOWN_SIZE, "", data))
}

////////////////////////////////////////////////////////////////////////////////

type norefBlobSource struct {
	BlobSource
}

var _ BlobSource = (*norefBlobSource)(nil)

func NoRefBlobSource(s BlobSource) BlobSource { return &norefBlobSource{s} }

func (norefBlobSource) Ref() error {
	return nil
}

func (norefBlobSource) Unref() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type norefBlobSink struct {
	BlobSink
}

var _ BlobSink = (*norefBlobSink)(nil)

func NoRefBlobSink(s BlobSink) BlobSink { return &norefBlobSink{s} }

func (norefBlobSink) Ref() error {
	return nil
}

func (norefBlobSink) Unref() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type norefBlobCache struct {
	BlobCache
}

var _ BlobCache = (*norefBlobCache)(nil)

func NoRefBlobCache(s BlobCache) BlobCache { return &norefBlobCache{s} }

func (norefBlobCache) Ref() error {
	return nil
}

func (norefBlobCache) Unref() error {
	return nil
}
