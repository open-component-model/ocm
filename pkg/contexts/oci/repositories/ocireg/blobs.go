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

package ocireg

import (
	"sync"

	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/docker/resolve"
	"github.com/open-component-model/ocm/pkg/errors"
)

type BlobContainer interface {
	GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error)
	AddBlob(blob cpi.BlobAccess) (int64, digest.Digest, error)
	Unref() error
}

type blobContainer struct {
	accessio.StaticAllocatable
	fetcher resolve.Fetcher
	pusher  resolve.Pusher
	mime    string
}

type BlobContainers struct {
	lock    sync.Mutex
	cache   accessio.BlobCache
	fetcher resolve.Fetcher
	pusher  resolve.Pusher
	mimes   map[string]BlobContainer
}

func NewBlobContainers(ctx cpi.Context, fetcher remotes.Fetcher, pusher resolve.Pusher) *BlobContainers {
	return &BlobContainers{
		cache:   cacheattr.Get(ctx),
		fetcher: fetcher,
		pusher:  pusher,
		mimes:   map[string]BlobContainer{},
	}
}

func (c *BlobContainers) Get(mime string) BlobContainer {
	c.lock.Lock()
	defer c.lock.Unlock()

	found := c.mimes[mime]
	if found == nil {
		found = NewBlobContainer(c.cache, mime, c.fetcher, c.pusher)
		c.mimes[mime] = found
	}
	return found
}

func (c *BlobContainers) Release() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	list := errors.ErrListf("releasing mime block caches")
	for _, b := range c.mimes {
		list.Add(b.Unref())
	}
	return list.Result()
}

func newBlobContainer(mime string, fetcher resolve.Fetcher, pusher resolve.Pusher) *blobContainer {
	return &blobContainer{
		mime:    mime,
		fetcher: fetcher,
		pusher:  pusher,
	}
}

func NewBlobContainer(cache accessio.BlobCache, mime string, fetcher resolve.Fetcher, pusher resolve.Pusher) BlobContainer {
	c := newBlobContainer(mime, fetcher, pusher)

	if cache == nil {
		return c
	}
	r, err := accessio.CachedAccess(c, c, cache)
	if err != nil {
		panic(err)
	}
	return r
}

func (n *blobContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	// fmt.Printf("orig get %s %s\n",n.mime, digest)
	acc, err := NewDataAccess(n.fetcher, digest, n.mime, false)
	return accessio.BLOB_UNKNOWN_SIZE, acc, err
}

func (n *blobContainer) AddBlob(blob cpi.BlobAccess) (int64, digest.Digest, error) {
	err := push(dummyContext, n.pusher, blob)
	if err != nil {
		return accessio.BLOB_UNKNOWN_SIZE, accessio.BLOB_UNKNOWN_DIGEST, err
	}
	return blob.Size(), blob.Digest(), err
}
