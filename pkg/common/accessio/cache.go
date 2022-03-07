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
	"sync"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/marstr/guid"
	"github.com/opencontainers/go-digest"
)

type BlobCache interface {
	GetBlob(mime string, digest digest.Digest) (BlobAccess, error)
	AddBlob(blob BlobAccess) (int64, digest.Digest, error)
	Close() error
}

type blobCache struct {
	lock     sync.RWMutex
	refcount int
	cache    vfs.FileSystem
}

func NewDefaultBlobCache() (BlobCache, error) {
	fs, err := osfs.NewTempFileSystem()
	if err != nil {
		return nil, err
	}
	return &blobCache{
		cache: fs,
	}, nil
}

func (c *blobCache) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := vfs.Cleanup(c.cache)
	c.cache = nil
	return err
}

func (c *blobCache) GetBlob(mime string, digest digest.Digest) (BlobAccess, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.cache == nil {
		return nil, ErrClosed
	}

	path := common.DigestToFileName(digest)
	if ok, err := vfs.Exists(c.cache, path); ok || err != nil {
		return BlobAccessForFile(mime, path, c.cache), nil
	}
	return nil, ErrBlobNotFound(digest)
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
