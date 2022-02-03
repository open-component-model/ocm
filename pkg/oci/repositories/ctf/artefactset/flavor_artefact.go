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
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

var ErrNoIndex = errors.New("manifest does not support access to subsequent artefacts")

type Artefact struct {
	lock     sync.RWMutex
	access   ArtefactSetContainer
	artefact *artdesc.Artefact
	handler  *BlobHandler
}

var _ cpi.ArtefactAccess = (*Artefact)(nil)

func NewArtefact(access ArtefactSetContainer, def ...*artdesc.Artefact) *Artefact {
	var artefact *artdesc.Artefact
	if len(def) == 0 || def[0] == nil {
		artefact = artdesc.New()
	} else {
		artefact = def[0]
	}
	a := &Artefact{
		access:   access,
		artefact: artefact,
	}
	a.handler = NewBlobHandler(access, a)
	return a
}

func (a *Artefact) IsClosed() bool {
	return a.access.IsClosed()
}

func (a *Artefact) IsReadOnly() bool {
	return a.access.IsReadOnly()
}

func (a *Artefact) Artefact() *artdesc.Artefact {
	a.lock.RLock()
	defer a.lock.RUnlock()
	if a.artefact.IsValid() {
		return a.artefact
	}
	return nil
}

func (a *Artefact) GetDescriptor() *artdesc.Artefact {
	if a.artefact.IsValid() {
		return a.artefact
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// from artdesc.Artefact

func (a *Artefact) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := a.artefact.GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return a.access.GetBlobDescriptor(digest)
}

func (a *Artefact) IsIndex() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.artefact.IsIndex()
}

func (a *Artefact) IsManifest() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.artefact.IsManifest()
}

func (a *Artefact) Index() (*artdesc.Index, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	idx := a.artefact.Index()
	if idx == nil {
		idx = artdesc.NewIndex()
		if err := a.artefact.SetIndex(idx); err != nil {
			return nil, err
		}
	}
	return idx, nil
}

func (a *Artefact) Manifest() (*artdesc.Manifest, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	m := a.artefact.Manifest()
	if m == nil {
		m = artdesc.NewManifest()
		if err := a.artefact.SetManifest(m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

////////////////////////////////////////////////////////////////////////////////
// from BlobHandler

func (a *Artefact) GetArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	if !a.IsIndex() {
		return nil, ErrNoIndex
	}
	return a.handler.GetArtefact(digest)
}

func (a *Artefact) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	return a.handler.GetBlob(digest)
}

func (a *Artefact) GetManifest(digest digest.Digest) (cpi.ManifestAccess, error) {
	if !a.IsIndex() {
		return nil, ErrNoIndex
	}
	return a.handler.GetManifest(digest)
}

func (a *Artefact) GetIndex(digest digest.Digest) (cpi.IndexAccess, error) {
	if !a.IsIndex() {
		return nil, ErrNoIndex
	}
	return a.handler.GetIndex(digest)
}

func (a *Artefact) NewArtefact(art ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	return a.handler.NewArtefact(art...)
}

////////////////////////////////////////////////////////////////////////////////

func (a *Artefact) AddBlob(access cpi.BlobAccess) error {
	return a.access.AddBlob(access)
}

func (a *Artefact) AddArtefact(art cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	idx, err := a.Index()
	if err != nil {
		return nil, err
	}
	return NewIndex(a.access, idx).AddArtefact(art, platform)
}

func (a *Artefact) AddLayer(blob cpi.BlobAccess, d *cpi.Descriptor) (int, error) {
	if a.IsClosed() {
		return -1, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return -1, accessio.ErrReadOnly
	}
	manifest, err := a.Manifest()
	if err != nil {
		return -1, err
	}
	return NewManifest(a.access, manifest).AddLayer(blob, d)
}
