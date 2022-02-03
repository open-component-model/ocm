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

package artefactset

import (
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

////////////////////////////////////////////////////////////////////////////////

type ArtefactSetAccess struct {
	base ArtefactSetContainer

	lock      sync.RWMutex
	blobinfos map[digest.Digest]*cpi.Descriptor

	*BlobHandler
}

func NewArtefactSetAccess(container ArtefactSetContainer) *ArtefactSetAccess {
	s := &ArtefactSetAccess{
		base:      container,
		blobinfos: map[digest.Digest]*cpi.Descriptor{},
	}
	s.BlobHandler = NewBlobHandler(s, s)
	return s
}

func (a *ArtefactSetAccess) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *ArtefactSetAccess) IsClosed() bool {
	return a.base.IsClosed()
}

////////////////////////////////////////////////////////////////////////////////
// methods for BlobHandler

func (a *ArtefactSetAccess) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
	return a.base.GetBlobData(digest)
}

func (a *ArtefactSetAccess) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	d := a.GetBlobDescriptor(digest)
	if d != nil {
		data, err := a.GetBlobData(digest)
		if err != nil {
			return nil, err
		}
		return accessio.BlobAccessForDataAccess(d.Digest, d.Size, d.MediaType, data), nil
	}
	return nil, errors.ErrNotFound("blob", string(digest))
}

func (a *ArtefactSetAccess) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	a.lock.RLock()
	defer a.lock.RUnlock()

	d := a.blobinfos[digest]
	if d == nil {
		d = a.base.GetBlobDescriptor(digest)
	}
	return d
}

func (a *ArtefactSetAccess) AddArtefact(artefact cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	return a.base.AddArtefact(artefact, platform)
}

func (a *ArtefactSetAccess) AddBlob(blob cpi.BlobAccess) error {
	a.lock.RLock()
	defer a.lock.RUnlock()
	err := a.base.AddBlob(blob)
	if err != nil {
		return err
	}
	a.blobinfos[blob.Digest()] = artdesc.DefaultBlobDescriptor(blob)
	return nil
}
