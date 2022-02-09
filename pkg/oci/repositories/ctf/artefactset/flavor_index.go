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
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

type Index struct {
	artefactBase
	*BlobHandler // Index offers all: blobs, artefacts, manifests and indices
}

var _ cpi.IndexAccess = (*Index)(nil)

func NewIndex(access ArtefactSetContainer, defs ...*artdesc.Index) core.IndexAccess {
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

	i := &Index{
		artefactBase: artefactBase{
			access: access,
			state:  state,
		},
	}
	i.BlobHandler = NewBlobHandler(access, i)
	return i
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

func NewIndexForArtefact(a *Artefact) *Index {
	m := &Index{
		artefactBase: artefactBase{
			access: a.access,
			state:  &indexMapper{a.state},
		},
	}
	m.BlobHandler = NewBlobHandler(a.access, m)
	return m
}

func (i *Index) Blob() (accessio.BlobAccess, error) {
	blob, err := i.artefactBase.Blob()
	if err != nil {
		return nil, err
	}
	return accessio.BlobWithMimeType(artdesc.MediaTypeImageIndex, blob), nil
}

func (i *Index) IsManifest() bool {
	return false
}

func (i *Index) IsIndex() bool {
	return true
}

func (i *Index) Manifest() (*artdesc.Manifest, error) {
	return nil, errors.ErrInvalid()
}

func (i *Index) Index() (*artdesc.Index, error) {
	return i.GetDescriptor(), nil
}

func (i *Index) Artefact() *artdesc.Artefact {
	a := artdesc.New()
	_ = a.SetIndex(i.GetDescriptor())
	return a
}

func (i *Index) GetDescriptor() *artdesc.Index {
	return i.state.GetState().(*artdesc.Index)
}

func (i *Index) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := i.GetDescriptor().GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return i.container.GetBlobDescriptor(digest)
}

func (a *Index) AddArtefact(art cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	blob, err := a.container.AddArtefact(art, platform)
	if err != nil {
		return nil, err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	d := a.GetDescriptor()
	d.Manifests = append(d.Manifests, cpi.Descriptor{
		MediaType:   blob.MimeType(),
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    platform,
	})
	return blob, nil
}
