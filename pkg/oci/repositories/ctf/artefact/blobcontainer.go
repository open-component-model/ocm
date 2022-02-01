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

package artefact

import (
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type BlobSource interface {
	GetBlob(digest digest.Digest) (core.BlobAccess, error)
}

type BlobSourceFunction func(digest digest.Digest) (core.BlobAccess, error)

func (f BlobSourceFunction) GetBlob(digest digest.Digest) (core.BlobAccess, error) {
	return f(digest)
}

type BlobContainer struct {
	artefact *Artefact
	BlobSource
}

func NewBlobContainer(artefact *Artefact, src BlobSource) *BlobContainer {
	return &BlobContainer{artefact, src}
}

func (i *BlobContainer) GetIndex(digest digest.Digest) (core.IndexAccess, error) {
	blob, err := i.GetBlob(digest)
	if err != nil {
		return nil, err
	}
	if blob.MimeType() != ociv1.MediaTypeImageIndex {
		return nil, errors.ErrInvalid(cpi.KIND_MEDIATYPE, blob.MimeType())
	}

	data, err := blob.Get()
	if err != nil {
		return nil, err
	}
	d, err := artdesc.Decode(data)
	if err != nil {
		return nil, err
	}
	if !d.IsIndex() {
		return nil, errors.Newf("blob is no index")
	}
	return NewIndex(i.artefact, d.Index()), nil
}

func (i *BlobContainer) GetManifest(digest digest.Digest) (core.ManifestAccess, error) {
	blob, err := i.GetBlob(digest)
	if err != nil {
		return nil, err
	}

	if blob.MimeType() != ociv1.MediaTypeImageManifest {
		return nil, errors.ErrInvalid(cpi.KIND_MEDIATYPE, blob.MimeType())
	}
	data, err := blob.Get()
	if err != nil {
		return nil, err
	}
	d, err := artdesc.Decode(data)
	if err != nil {
		return nil, err
	}
	if !d.IsManifest() {
		return nil, errors.Newf("blob is no manifest")
	}
	return &Manifest{i.artefact, d.Manifest()}, nil
}
