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
	"github.com/containerd/containerd/remotes"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/opencontainers/go-digest"
)

type BlobContainer interface {
	GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error)
	AddBlob(blob cpi.BlobAccess) (int64, digest.Digest, error)
}

type blobContainer struct {
	accessio.StaticAllocatable
	fetcher remotes.Fetcher
	pusher  remotes.Pusher
}

func newBlobContainer(fetcher remotes.Fetcher, pusher remotes.Pusher) *blobContainer {
	return &blobContainer{
		fetcher: fetcher,
		pusher:  pusher,
	}
}

func NewBlobContainer(ctx cpi.Context, fetcher remotes.Fetcher, pusher remotes.Pusher) BlobContainer {
	c := newBlobContainer(fetcher, pusher)
	cache := cacheattr.Get(ctx)

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
	acc, err := NewDataAccess(n.fetcher, digest, "", false)
	return accessio.BLOB_UNKNOWN_SIZE, acc, err
}

func (n *blobContainer) AddBlob(blob cpi.BlobAccess) (int64, digest.Digest, error) {
	err := push(dummyContext, n.pusher, blob)
	if err != nil {
		return accessio.BLOB_UNKNOWN_SIZE, accessio.BLOB_UNKNOWN_DIGEST, err
	}
	return blob.Size(), blob.Digest(), err
}
