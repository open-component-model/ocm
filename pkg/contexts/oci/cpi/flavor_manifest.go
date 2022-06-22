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

package cpi

import (
	"compress/gzip"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ManifestImpl struct {
	artefactBase
}

var _ ManifestAccess = (*ManifestImpl)(nil)

func NewManifest(access ArtefactSetContainer, defs ...*artdesc.Manifest) (*ManifestImpl, error) {
	var def *artdesc.Manifest
	if len(defs) != 0 && defs[0] != nil {
		def = defs[0]
	}
	mode := accessobj.ACC_WRITABLE
	if access.IsReadOnly() {
		mode = accessobj.ACC_READONLY
	}
	state, err := accessobj.NewBlobStateForObject(mode, def, NewManifestStateHandler())
	if err != nil {
		panic("oops")
	}

	p, err := access.NewArtefactProvider(state)
	if err != nil {
		return nil, err
	}

	m := &ManifestImpl{
		artefactBase: artefactBase{
			container: access,
			state:     state,
			provider:  p,
		},
	}
	return m, nil
}

type manifestMapper struct {
	accessobj.State
}

var _ accessobj.State = (*manifestMapper)(nil)

func (m *manifestMapper) GetState() interface{} {
	return m.State.GetState().(*artdesc.Artefact).Manifest()
}
func (m *manifestMapper) GetOriginalState() interface{} {
	return m.State.GetOriginalState().(*artdesc.Artefact).Manifest()
}

func NewManifestForArtefact(a *ArtefactImpl) *ManifestImpl {
	m := &ManifestImpl{
		artefactBase: artefactBase{
			container: a.container,
			state:     &manifestMapper{a.state},
			provider:  a.provider,
		},
	}
	return m
}

func (m *ManifestImpl) AddBlob(access BlobAccess) error {
	return m.addBlob(access)
}

func (m *ManifestImpl) Manifest() (*artdesc.Manifest, error) {
	return m.GetDescriptor(), nil
}

func (m *ManifestImpl) Index() (*artdesc.Index, error) {
	return nil, errors.ErrInvalid()
}

func (m *ManifestImpl) Artefact() *artdesc.Artefact {
	a := artdesc.New()
	_ = a.SetManifest(m.GetDescriptor())
	return a
}

func (m *ManifestImpl) GetDescriptor() *artdesc.Manifest {
	return m.state.GetState().(*artdesc.Manifest)
}

func (m *ManifestImpl) GetBlobDescriptor(digest digest.Digest) *Descriptor {
	d := m.GetDescriptor().GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return m.container.GetBlobDescriptor(digest)
}

func (m *ManifestImpl) GetConfigBlob() (BlobAccess, error) {
	if m.GetDescriptor().Config.Digest == "" {
		return nil, nil
	}
	return m.GetBlob(m.GetDescriptor().Config.Digest)
}

func (m *ManifestImpl) GetBlob(digest digest.Digest) (BlobAccess, error) {
	d := m.GetBlobDescriptor(digest)
	if d != nil {
		size, data, err := m.provider.GetBlobData(digest)
		if err != nil {
			return nil, err
		}
		err = AdjustSize(d, size)
		if err != nil {
			return nil, err
		}
		return accessio.BlobAccessForDataAccess(d.Digest, d.Size, d.MediaType, data), nil
	}
	return nil, ErrBlobNotFound(digest)
}

func (m *ManifestImpl) SetConfigBlob(blob BlobAccess, d *artdesc.Descriptor) error {
	if d == nil {
		d = artdesc.DefaultBlobDescriptor(blob)
	}
	err := m.AddBlob(blob)
	if err != nil {
		return err
	}
	m.GetDescriptor().Config = *d
	return nil
}

func (m *ManifestImpl) AddLayer(blob BlobAccess, d *artdesc.Descriptor) (int, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if d == nil {
		d = &artdesc.Descriptor{}
	}
	d.Digest = blob.Digest()
	d.Size = blob.Size()
	if d.MediaType == "" {
		d.MediaType = blob.MimeType()
		if d.MediaType == "" {
			d.MediaType = artdesc.MediaTypeImageLayer
			r, err := blob.Reader()
			if err != nil {
				return -1, err
			}
			defer r.Close()
			zr, err := gzip.NewReader(r)
			if err == nil {
				err = zr.Close()
				if err == nil {
					d.MediaType = artdesc.MediaTypeImageLayerGzip
				}
			}
		}
	}

	err := m.container.AddBlob(blob)
	if err != nil {
		return -1, err
	}

	manifest := m.GetDescriptor()
	manifest.Layers = append(manifest.Layers, *d)
	return len(manifest.Layers) - 1, nil
}
