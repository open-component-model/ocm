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
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/core"
	"github.com/open-component-model/ocm/pkg/errors"
)

type IndexImpl struct {
	artefactBase
}

var _ IndexAccess = (*IndexImpl)(nil)

func NewIndex(access ArtefactSetContainer, defs ...*artdesc.Index) (core.IndexAccess, error) {
	var def *artdesc.Index
	if len(defs) != 0 && defs[0] != nil {
		def = defs[0]
	}
	mode := accessobj.ACC_WRITABLE
	if access.IsReadOnly() {
		mode = accessobj.ACC_READONLY
	}
	state, err := accessobj.NewBlobStateForObject(mode, def, NewIndexStateHandler())
	if err != nil {
		panic("oops")
	}

	p, err := access.NewArtefactProvider(state)
	if err != nil {
		return nil, err
	}
	i := &IndexImpl{
		artefactBase: artefactBase{
			container: access,
			provider:  p,
			state:     state,
		},
	}
	return i, nil
}

type indexMapper struct {
	accessobj.State
}

var _ accessobj.State = (*indexMapper)(nil)

func (m *indexMapper) GetState() interface{} {
	return m.State.GetState().(*artdesc.Artefact).Index()
}
func (m *indexMapper) GetOriginalState() interface{} {
	return m.State.GetOriginalState().(*artdesc.Artefact).Index()
}

func NewIndexForArtefact(a *ArtefactImpl) *IndexImpl {
	m := &IndexImpl{
		artefactBase: artefactBase{
			container: a.container,
			provider:  a.provider,
			state:     &indexMapper{a.state},
		},
	}
	return m
}

func (a *IndexImpl) NewArtefact(art ...*artdesc.Artefact) (ArtefactAccess, error) {
	return a.newArtefact(art...)
}

func (i *IndexImpl) AddBlob(blob core.BlobAccess) error {
	return i.provider.AddBlob(blob)
}

func (i *IndexImpl) Manifest() (*artdesc.Manifest, error) {
	return nil, errors.ErrInvalid()
}

func (i *IndexImpl) Index() (*artdesc.Index, error) {
	return i.GetDescriptor(), nil
}

func (i *IndexImpl) Artefact() *artdesc.Artefact {
	a := artdesc.New()
	_ = a.SetIndex(i.GetDescriptor())
	return a
}

func (i *IndexImpl) GetDescriptor() *artdesc.Index {
	return i.state.GetState().(*artdesc.Index)
}

func (i *IndexImpl) GetBlobDescriptor(digest digest.Digest) *Descriptor {
	d := i.GetDescriptor().GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return i.provider.GetBlobDescriptor(digest)
}

func (i *IndexImpl) GetBlob(digest digest.Digest) (core.BlobAccess, error) {
	d := i.GetBlobDescriptor(digest)
	if d != nil {
		size, data, err := i.provider.GetBlobData(digest)
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

func (i *IndexImpl) GetArtefact(digest digest.Digest) (core.ArtefactAccess, error) {
	for _, d := range i.GetDescriptor().Manifests {
		if d.Digest == digest {
			return i.provider.GetArtefact(digest)
		}
	}
	return nil, errors.ErrNotFound(KIND_OCIARTEFACT, digest.String())
}

func (i *IndexImpl) GetIndex(digest digest.Digest) (core.IndexAccess, error) {
	a, err := i.GetArtefact(digest)
	if err != nil {
		return nil, err
	}
	if idx, err := a.Index(); err == nil {
		return NewIndex(i.container, idx)
	}
	return nil, errors.New("no index")
}

func (i *IndexImpl) GetManifest(digest digest.Digest) (core.ManifestAccess, error) {
	a, err := i.GetArtefact(digest)
	if err != nil {
		return nil, err
	}
	if m, err := a.Manifest(); err == nil {
		return NewManifest(i.container, m)
	}
	return nil, errors.New("no manifest")
}

func (a *IndexImpl) AddArtefact(art Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	blob, err := a.provider.AddArtefact(art)
	if err != nil {
		return nil, err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	d := a.GetDescriptor()
	d.Manifests = append(d.Manifests, Descriptor{
		MediaType:   blob.MimeType(),
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    platform,
	})
	return blob, nil
}
