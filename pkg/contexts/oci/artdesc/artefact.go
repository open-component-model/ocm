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

package artdesc

import (
	"encoding/json"

	"github.com/containerd/containerd/images"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc/helper"
	"github.com/open-component-model/ocm/pkg/errors"
)

const SchemeVersion = helper.SchemeVersion

const (
	MediaTypeImageManifest  = ociv1.MediaTypeImageManifest
	MediaTypeImageIndex     = ociv1.MediaTypeImageIndex
	MediaTypeImageLayer     = ociv1.MediaTypeImageLayer
	MediaTypeImageLayerGzip = ociv1.MediaTypeImageLayerGzip

	MediaTypeDockerSchema2Manifest     = images.MediaTypeDockerSchema2Manifest
	MediaTypeDockerSchema2ManifestList = images.MediaTypeDockerSchema2ManifestList
)

var legacy = false

type Descriptor = ociv1.Descriptor
type Platform = ociv1.Platform

type BlobDescriptorSource interface {
	GetBlobDescriptor(digest.Digest) *Descriptor
	MimeType() string
	IsValid() bool
}

// Artefact is the unified representation of an OCI artefact
// according to https://github.com/opencontainers/image-spec/blob/main/manifest.md
// It is either an image manifest or an image index manifest (fat image)
type Artefact struct {
	manifest *Manifest
	index    *Index
}

var (
	_ BlobDescriptorSource = (*Artefact)(nil)
	_ json.Marshaler       = (*Artefact)(nil)
	_ json.Unmarshaler     = (*Artefact)(nil)
)

func New() *Artefact {
	return &Artefact{}
}

func NewManifestArtefact() *Artefact {
	a := New()
	a.SetManifest(NewManifest())
	return a
}

func NewIndexArtefact() *Artefact {
	a := New()
	a.SetIndex(NewIndex())
	return a
}

func (d *Artefact) MimeType() string {
	if d.IsIndex() {
		return d.index.MimeType()
	}
	if d.IsManifest() {
		return d.manifest.MimeType()
	}
	return ""
}

func (d *Artefact) SetManifest(m *Manifest) error {
	if d.IsIndex() || d.IsManifest() {
		return errors.Newf("artefact descriptor already instantiated")
	}
	d.manifest = m
	return nil
}

func (d *Artefact) SetIndex(i *Index) error {
	if d.IsIndex() || d.IsManifest() {
		return errors.Newf("artefact descriptor already instantiated")
	}
	d.index = i
	return nil
}

func (d *Artefact) IsValid() bool {
	return d.manifest != nil || d.index != nil
}

func (d *Artefact) IsManifest() bool {
	return d.manifest != nil
}

func (d *Artefact) IsIndex() bool {
	return d.index != nil
}

func (d *Artefact) Index() *Index {
	return d.index
}

func (d *Artefact) Manifest() *Manifest {
	return d.manifest
}

func (d *Artefact) ToBlobAccess() (accessio.BlobAccess, error) {
	if d.IsManifest() {
		return d.manifest.ToBlobAccess()
	}
	if d.IsIndex() {
		return d.index.ToBlobAccess()
	}
	return nil, errors.ErrInvalid("artefact descriptor")
}

func (d *Artefact) GetBlobDescriptor(digest digest.Digest) *Descriptor {
	if d.IsManifest() {
		return d.Manifest().GetBlobDescriptor(digest)
	}
	if d.IsIndex() {
		return d.Index().GetBlobDescriptor(digest)
	}
	return nil
}

func (d Artefact) MarshalJSON() ([]byte, error) {
	if d.manifest != nil {
		d.manifest.MediaType = ArtefactMimeType(d.manifest.MediaType, ociv1.MediaTypeImageManifest, legacy)
		return json.Marshal(d.manifest)
	}
	if d.index != nil {
		d.index.MediaType = ArtefactMimeType(d.index.MediaType, ociv1.MediaTypeImageIndex, legacy)
		return json.Marshal(d.index)
	}
	return []byte("null"), nil
}

func (d *Artefact) UnmarshalJSON(data []byte) error {

	if string(data) == "null" {
		return nil
	}
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
		d.manifest = (*Manifest)(m.AsManifest())
		d.index = nil
	} else {
		d.index = (*Index)(m.AsIndex())
		d.manifest = nil
	}
	return nil
}

func Decode(data []byte) (*Artefact, error) {
	var d Artefact

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func Encode(d *Artefact) ([]byte, error) {
	return json.Marshal(d)
}
