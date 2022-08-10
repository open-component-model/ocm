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

type CachedBlob struct {
	lock sync.Mutex
	mime string

	cache     *tmpcache.Attribute
	path      string
	digest    digest.Digest
	size      int64
	source    accessio.DataAccess
	effective accessio.BlobAccess
}

var _ accessio.BlobAccess = (*CachedBlob)(nil)

func NewCachedBlob(ctx datacontext.Context, mime string, src accessio.DataAccess) accessio.BlobAccess {
	return &CachedBlob{
		source: src,
		mime:   mime,
		cache:  tmpcache.Get(ctx),
	}
}

func (c *CachedBlob) setup() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.effective != nil {
		return nil
	}
	file, err := c.cache.CreateTempFile("blob*")
	if err != nil {
		return err
	}
	c.path = file.Name()

	r, err := c.source.Reader()
	dr := accessio.NewDefaultDigestReader(r)
	if err != nil {
		return err
	}
	defer r.Close()
	defer file.Close()
	_, err = io.Copy(file, dr)
	if err != nil {
		return err
	}
	c.effective = accessio.BlobAccessForFile(c.mime, c.path, c.cache.Filesystem)
	c.digest = dr.Digest()
	c.size = dr.Size()
	return err
}

func (c *CachedBlob) Get() ([]byte, error) {
	err := c.setup()
	if err != nil {
		return nil, err
	}
	return c.effective.Get()
}

func (c *CachedBlob) Reader() (io.ReadCloser, error) {
	err := c.setup()
	if err != nil {
		return nil, err
	}
	return c.effective.Reader()
}

func (c *CachedBlob) Close() error {
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

func (c *CachedBlob) Digest() digest.Digest {
	err := c.setup()
	if err != nil {
		return ""
	}
	return c.digest
}

func (c *CachedBlob) MimeType() string {
	panic("implement me")
}

func (c *CachedBlob) DigestKnown() bool {
	return c.effective != nil
}

func (c *CachedBlob) Size() int64 {
	err := c.setup()
	if err != nil {
		return -1
	}
	return c.size
}
