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
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

type BlobDescriptorSource interface {
	GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor
}

type BlobHandler struct {
	set ArtefactSetContainer
	BlobDescriptorSource
}

func NewBlobHandler(set ArtefactSetContainer, src BlobDescriptorSource) *BlobHandler {
	return &BlobHandler{set, src}
}

func (i *BlobHandler) AddBlob(blob cpi.BlobAccess) error {
	return i.set.AddBlob(blob)
}

func (i *BlobHandler) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	d := i.GetBlobDescriptor(digest)
	if d != nil {
		data, err := i.set.GetBlobData(digest)
		if err != nil {
			return nil, err
		}
		return accessio.BlobAccessForDataAccess(d.Digest, d.Size, d.MediaType, data), nil
	}
	return nil, cpi.ErrBlobNotFound(digest)
}

func (i *BlobHandler) GetArtefact(digest digest.Digest) (*Artefact, error) {
	return i.set.GetArtefact(digest)
}

func (i *BlobHandler) GetIndex(digest digest.Digest) (core.IndexAccess, error) {
	a, err := i.GetArtefact(digest)
	if err != nil {
		return nil, err
	}
	if idx, err := a.Index(); err == nil {
		return NewIndex(i.set, idx), nil
	}
	return nil, errors.New("no index")
}

func (i *BlobHandler) GetManifest(digest digest.Digest) (core.ManifestAccess, error) {
	a, err := i.GetArtefact(digest)
	if err != nil {
		return nil, err
	}
	if m, err := a.Manifest(); err == nil {
		return NewManifest(i.set, m), nil
	}
	return nil, errors.New("no manifest")
}
