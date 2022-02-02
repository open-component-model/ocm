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
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type BlobSource interface {
	GetBlob(digest digest.Digest) (cpi.BlobAccess, error)
}

type BlobContainer struct {
	set *ArtefactSet
	BlobSource
}

func NewBlobContainer(artefact *ArtefactSet, src BlobSource) *BlobContainer {
	return &BlobContainer{artefact, src}
}

func (i *BlobContainer) AddBlob(blob cpi.BlobAccess) error {
	return i.set.AddBlob(blob)
}

func (i *BlobContainer) getArtefact(blob cpi.BlobAccess) (*artdesc.Artefact, error) {
	data, err := blob.Get()
	if err != nil {
		return nil, err
	}
	return artdesc.Decode(data)
}

func (i *BlobContainer) GetArtefact(digest digest.Digest) (*Artefact, error) {
	blob, err := i.GetBlob(digest)
	if err != nil {
		return nil, err
	}

	d, err := i.getArtefact(blob)
	if err != nil {
		return nil, err
	}
	return NewArtefact(i.set, d), nil
}

func (i *BlobContainer) GetIndex(digest digest.Digest) (core.IndexAccess, error) {
	blob, err := i.GetBlob(digest)
	if err != nil {
		return nil, err
	}
	if blob.MimeType() != ociv1.MediaTypeImageIndex {
		return nil, errors.ErrInvalid(cpi.KIND_MEDIATYPE, blob.MimeType())
	}

	d, err := i.getArtefact(blob)
	if !d.IsIndex() {
		return nil, errors.Newf("blob is no index")
	}
	return NewIndex(i.set, d.Index()), nil
}

func (i *BlobContainer) GetManifest(digest digest.Digest) (core.ManifestAccess, error) {
	blob, err := i.GetBlob(digest)
	if err != nil {
		return nil, err
	}

	if blob.MimeType() != ociv1.MediaTypeImageManifest {
		return nil, errors.ErrInvalid(cpi.KIND_MEDIATYPE, blob.MimeType())
	}
	d, err := i.getArtefact(blob)
	if err != nil {
		return nil, err
	}
	if !d.IsManifest() {
		return nil, errors.Newf("blob is no manifest")
	}
	return &Manifest{i.set, d.Manifest()}, nil
}
