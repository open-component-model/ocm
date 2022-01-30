// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package artdesc

import (
	"encoding/json"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc/helper"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const SchemeVersion = helper.SchemeVersion

const (
	MediaTypeImageManifest  = ociv1.MediaTypeImageManifest
	MediaTypeImageIndex     = ociv1.MediaTypeImageIndex
	MediaTypeImageLayer     = ociv1.MediaTypeImageLayer
	MediaTypeImageLayerGzip = ociv1.MediaTypeImageLayerGzip
)

type Manifest = ociv1.Manifest
type Index = ociv1.Index
type Descriptor = ociv1.Descriptor
type Platform = ociv1.Platform

type ArtefactDescriptor struct {
	manifest *Manifest
	index    *Index
}

var _ json.Marshaler = &ArtefactDescriptor{}
var _ json.Unmarshaler = &ArtefactDescriptor{}

func New() *ArtefactDescriptor {
	return &ArtefactDescriptor{}
}

func (d *ArtefactDescriptor) SetManifest(m *Manifest) error {
	if d.IsIndex() || d.IsManifest() {
		return errors.Newf("artefact descriptor already instantiated")
	}
	d.manifest = m
	return nil
}

func (d *ArtefactDescriptor) SetIndex(i *Index) error {
	if d.IsIndex() || d.IsManifest() {
		return errors.Newf("artefact descriptor already instantiated")
	}
	d.index = i
	return nil
}

func (d *ArtefactDescriptor) IsValid() bool {
	return d.manifest != nil || d.index != nil
}

func (d *ArtefactDescriptor) IsManifest() bool {
	return d.manifest != nil
}

func (d *ArtefactDescriptor) IsIndex() bool {
	return d.index != nil
}

func (d *ArtefactDescriptor) Index() *Index {
	return d.index
}

func (d *ArtefactDescriptor) Manifest() *Manifest {
	return d.manifest
}

func (d *ArtefactDescriptor) ToBlobAccess() (accessio.BlobAccess, error) {
	if d.IsManifest() {
		return BlobAccessForManifest(d.manifest)
	}
	if d.IsIndex() {
		return BlobAccessForIndex(d.index)
	}
	return nil, errors.ErrInvalid("artefact descriptor")
}

func (d *ArtefactDescriptor) GetBlobDescriptor(digest digest.Digest) *Descriptor {
	if d.IsManifest() {
		return GetBlobDescriptorFromManifest(digest, d.Manifest())
	}
	if d.IsIndex() {
		return GetBlobDescriptorFromIndex(digest, d.Index())
	}
	return nil
}

func (d ArtefactDescriptor) MarshalJSON() ([]byte, error) {
	if d.manifest != nil {
		d.manifest.MediaType = ociv1.MediaTypeImageManifest
		return json.Marshal(d.manifest)
	}
	if d.index != nil {
		d.manifest.MediaType = ociv1.MediaTypeImageIndex
		return json.Marshal(d.index)
	}
	return []byte("{}"), nil
}

func (d ArtefactDescriptor) UnmarshalJSON(data []byte) error {
	var m helper.GenericDescriptor

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	err = m.Validate()
	if err != nil {
		return err
	}
	if m.IsManifest() {
		d.manifest = m.AsManifest()
		d.index = nil
	}
	d.index = m.AsIndex()
	d.manifest = nil
	return nil
}

func Decode(data []byte) (*ArtefactDescriptor, error) {
	var d ArtefactDescriptor

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func Encode(d *ArtefactDescriptor) ([]byte, error) {
	return json.Marshal(d)
}

func NewIndex() *Index {
	return &Index{
		Versioned:   specs.Versioned{SchemeVersion},
		MediaType:   MediaTypeImageIndex,
		Manifests:   nil,
		Annotations: nil,
	}
}

func NewManifest() *Manifest {
	return &Manifest{
		Versioned:   specs.Versioned{SchemeVersion},
		MediaType:   MediaTypeImageManifest,
		Layers:      nil,
		Annotations: nil,
	}
}

func DefaultBlobDescriptor(blob accessio.BlobAccess) *Descriptor {
	return &Descriptor{
		MediaType:   blob.MimeType(),
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    nil,
	}
}

func BlobAccessForManifest(m *Manifest) (accessio.BlobAccess, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return accessio.BlobAccessForData(MediaTypeImageManifest, data), nil
}

func BlobAccessForIndex(m *Index) (accessio.BlobAccess, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return accessio.BlobAccessForData(MediaTypeImageIndex, data), nil
}
