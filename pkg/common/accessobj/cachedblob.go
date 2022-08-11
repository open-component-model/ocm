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

package accessobj

import (
	"io"
	"sync"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/tmpcache"
)

type CachedBlobAccess struct {
	lock sync.Mutex
	mime string

	cache     *tmpcache.Attribute
	path      string
	digest    digest.Digest
	size      int64
	source    accessio.DataWriter
	effective accessio.BlobAccess
}

var _ accessio.BlobAccess = (*CachedBlobAccess)(nil)

func CachedBlobAccessForWriter(ctx datacontext.Context, mime string, src accessio.DataWriter) accessio.BlobAccess {
	return &CachedBlobAccess{
		source: src,
		mime:   mime,
		cache:  tmpcache.Get(ctx),
	}
}

func CachedBlobAccessForDataAccess(ctx datacontext.Context, mime string, src accessio.DataAccess) accessio.BlobAccess {
	return CachedBlobAccessForWriter(ctx, mime, accessio.NewDataAccessWriter(src))
}

func (c *CachedBlobAccess) setup() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.effective != nil {
		return nil
	}
	file, err := c.cache.CreateTempFile("blob*")
	if err != nil {
		return err
	}
	defer file.Close()
	c.path = file.Name()

	c.size, c.digest, err = c.source.WriteTo(file)
	if err != nil {
		return err
	}
	c.effective = accessio.BlobAccessForFile(c.mime, c.path, c.cache.Filesystem)
	return err
}

func (c *CachedBlobAccess) Get() ([]byte, error) {
	err := c.setup()
	if err != nil {
		return nil, err
	}
	return c.effective.Get()
}

func (c *CachedBlobAccess) Reader() (io.ReadCloser, error) {
	err := c.setup()
	if err != nil {
		return nil, err
	}
	return c.effective.Reader()
}

func (c *CachedBlobAccess) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	if c.effective != nil {
		c.effective.Close()
		err = c.cache.Filesystem.Remove(c.path)
	}
	c.effective = nil
	return err
}

func (c *CachedBlobAccess) Digest() digest.Digest {
	err := c.setup()
	if err != nil {
		return accessio.BLOB_UNKNOWN_DIGEST
	}
	if c.digest == accessio.BLOB_UNKNOWN_DIGEST {
		return c.effective.Digest()
	}
	return c.digest
}

func (c *CachedBlobAccess) MimeType() string {
	return c.mime
}

func (c *CachedBlobAccess) DigestKnown() bool {
	return c.effective != nil
}

func (c *CachedBlobAccess) Size() int64 {
	err := c.setup()
	if err != nil {
		return accessio.BLOB_UNKNOWN_SIZE
	}
	if c.size == accessio.BLOB_UNKNOWN_SIZE {
		return c.effective.Size()
	}
	return c.size
}
