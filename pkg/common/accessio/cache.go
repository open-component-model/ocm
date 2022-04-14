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
	"io"
	"os"
	"sync"

	"github.com/mandelsoft/vfs/pkg/osfs"
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
}

type blobCache struct {
	lock     sync.RWMutex
	cache    vfs.FileSystem
	refcount int
}

func NewDefaultBlobCache() (BlobCache, error) {
	fs, err := osfs.NewTempFileSystem()
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
	return -1, nil, err
}

func (c *blobCache) AddBlob(blob BlobAccess) (int64, digest.Digest, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cache == nil {
		return -1, "", ErrClosed
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
		return 0, "", errors.Wrapf(err, "cannot get blob content")
	}
	defer br.Close()

	var reader io.Reader
	if !blob.DigestKnown() {
		digester = NewDefaultDigestReader(reader)
		reader = digester
	} else {
		reader = br
	}

	writer, err := c.cache.Create(path)
	if err != nil {
		return -1, "", errors.Wrapf(err, "cannot create blob file in cache")
	}
	defer writer.Close()
	size, err := io.Copy(writer, reader)
	if err != nil {
		c.cache.Remove(path)
		return -1, "", err
	}
	if digester != nil {
		err = c.cache.Rename(path, common.DigestToFileName(digester.Digest()))
		if err != nil {
			c.cache.Remove(path)
		}
		return size, digester.Digest(), err
	}
	return size, blob.Digest(), nil
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
	return -1, nil, ErrBlobNotFound(digest)
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
