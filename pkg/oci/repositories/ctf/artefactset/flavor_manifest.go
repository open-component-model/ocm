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
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

type Manifest struct {
	access   ArtefactSetContainer
	manifest *artdesc.Manifest
	handler  *BlobHandler // not inherited because only blob access should be offered
}

var _ cpi.ManifestAccess = (*Manifest)(nil)

func NewManifest(access ArtefactSetContainer, manifest *artdesc.Manifest) *Manifest {
	m := &Manifest{
		access:   access,
		manifest: manifest,
	}
	m.handler = NewBlobHandler(access, m)
	return m
}

func (m *Manifest) GetDescriptor() *artdesc.Manifest {
	return m.manifest
}

func (m *Manifest) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := m.manifest.GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return m.access.GetBlobDescriptor(digest)
}

func (i *Manifest) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	return i.handler.GetBlob(digest)
}

func (m *Manifest) AddBlob(blob cpi.BlobAccess) error {
	return m.access.AddBlob(blob)
}
