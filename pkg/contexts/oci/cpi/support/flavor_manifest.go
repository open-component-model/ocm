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

package support

import (
	"compress/gzip"
	"fmt"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ManifestImpl struct {
	*artefactBase
}

var _ cpi.ManifestAccess = (*ManifestImpl)(nil)

type manifestMapper struct {
	accessobj.State
}

var _ accessobj.State = (*manifestMapper)(nil)

func (m *manifestMapper) GetState() interface{} {
	return m.State.GetState().(*artdesc.Artefact).Manifest()
}
func (m *manifestMapper) GetOriginalState() (interface{}, error) {
	state, err := m.State.GetOriginalState()
	if err != nil {
		return nil, fmt.Errorf("failed to return original state: %w", err)
	}
	return state.(*artdesc.Artefact).Manifest(), nil
}

func NewManifestForArtefact(a *ArtefactImpl) *ManifestImpl {
	m := &ManifestImpl{
		artefactBase: newArtefactBase(a.view, a.container, &manifestMapper{a.state}),
	}
	return m
}

func (m *ManifestImpl) AddBlob(access cpi.BlobAccess) error {
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

func (m *ManifestImpl) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := m.GetDescriptor().GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return m.container.GetBlobDescriptor(digest)
}

func (m *ManifestImpl) GetConfigBlob() (cpi.BlobAccess, error) {
	if m.GetDescriptor().Config.Digest == "" {
		return nil, nil
	}
	return m.GetBlob(m.GetDescriptor().Config.Digest)
}

func (m *ManifestImpl) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	d := m.GetBlobDescriptor(digest)
	if d != nil {
		size, data, err := m.container.GetBlobData(digest)
		if err != nil {
			return nil, err
		}
		err = AdjustSize(d, size)
		if err != nil {
			return nil, err
		}
		return accessio.BlobAccessForDataAccess(d.Digest, d.Size, d.MediaType, data), nil
	}
	return nil, cpi.ErrBlobNotFound(digest)
}

func (m *ManifestImpl) SetConfigBlob(blob cpi.BlobAccess, d *artdesc.Descriptor) error {
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

func (m *ManifestImpl) AddLayer(blob cpi.BlobAccess, d *artdesc.Descriptor) (int, error) {
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
