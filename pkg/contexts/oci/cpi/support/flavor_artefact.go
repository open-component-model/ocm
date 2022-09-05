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
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/core"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

var ErrNoIndex = errors.New("manifest does not support access to subsequent artefacts")

type ArtefactImpl struct {
	*artefactBase
}

var _ cpi.ArtefactAccess = (*ArtefactImpl)(nil)

func NewArtefactForBlob(container ArtefactSetContainerImpl, blob accessio.BlobAccess) (cpi.ArtefactAccess, error) {
	mode := accessobj.ACC_WRITABLE
	if container.IsReadOnly() {
		mode = accessobj.ACC_READONLY
	}
	state, err := accessobj.NewBlobStateForBlob(mode, blob, cpi.NewArtefactStateHandler())
	if err != nil {
		return nil, err
	}

	return newArtefactImpl(container, state)
}

func NewArtefact(container ArtefactSetContainerImpl, defs ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	var def *artdesc.Artefact
	if len(defs) != 0 && defs[0] != nil {
		def = defs[0]
	}
	mode := accessobj.ACC_WRITABLE
	if container.IsReadOnly() {
		mode = accessobj.ACC_READONLY
	}
	state, err := accessobj.NewBlobStateForObject(mode, def, cpi.NewArtefactStateHandler())
	if err != nil {
		panic("oops: " + err.Error())
	}

	return newArtefactImpl(container, state)
}

func newArtefactImpl(container ArtefactSetContainerImpl, state accessobj.State) (cpi.ArtefactAccess, error) {
	v, err := container.View()
	if err != nil {
		return nil, err
	}
	a := &ArtefactImpl{
		artefactBase: newArtefactBase(v, container, state),
	}
	return a, nil
}

func (a *ArtefactImpl) Close() error {
	return a.view.Close()
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (a *ArtefactImpl) AddBlob(access cpi.BlobAccess) error {
	return a.addBlob(access)
}

func (a *ArtefactImpl) NewArtefact(art ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	if !a.IsIndex() {
		return nil, ErrNoIndex
	}
	return a.newArtefact(art...)
}

////////////////////////////////////////////////////////////////////////////////

func (a *ArtefactImpl) Artefact() *artdesc.Artefact {
	return a.GetDescriptor()
}

func (a *ArtefactImpl) GetDescriptor() *artdesc.Artefact {
	d := a.state.GetState().(*artdesc.Artefact)
	if d.IsValid() {
		return d
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// from artdesc.Artefact

func (a *ArtefactImpl) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := a.GetDescriptor().GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return a.container.GetBlobDescriptor(digest)
}

func (a *ArtefactImpl) Index() (*artdesc.Index, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	d := a.state.GetState().(*artdesc.Artefact)
	idx := d.Index()
	if idx == nil {
		idx = artdesc.NewIndex()
		if err := d.SetIndex(idx); err != nil {
			return nil, errors.Newf("artefact is manifest")
		}
	}
	return idx, nil
}

func (a *ArtefactImpl) Manifest() (*artdesc.Manifest, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	d := a.state.GetState().(*artdesc.Artefact)
	m := d.Manifest()
	if m == nil {
		m = artdesc.NewManifest()
		if err := d.SetManifest(m); err != nil {
			return nil, errors.Newf("artefact is index")
		}
	}
	return m, nil
}

func (a *ArtefactImpl) ManifestAccess() core.ManifestAccess {
	a.lock.Lock()
	defer a.lock.Unlock()
	d := a.state.GetState().(*artdesc.Artefact)
	m := d.Manifest()
	if m == nil {
		m = artdesc.NewManifest()
		if err := d.SetManifest(m); err != nil {
			return nil
		}
	}
	return NewManifestForArtefact(a)
}

func (a *ArtefactImpl) IndexAccess() core.IndexAccess {
	a.lock.Lock()
	defer a.lock.Unlock()
	d := a.state.GetState().(*artdesc.Artefact)
	i := d.Index()
	if i == nil {
		i = artdesc.NewIndex()
		if err := d.SetIndex(i); err != nil {
			return nil
		}
	}
	return NewIndexForArtefact(a)
}

func (a *ArtefactImpl) GetArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	if !a.IsIndex() {
		return nil, ErrNoIndex
	}
	return a.container.GetArtefact("@" + digest.String())
}

func (a *ArtefactImpl) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return a.container.GetBlobData(digest)
}

func (a *ArtefactImpl) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	d := a.GetBlobDescriptor(digest)
	if d != nil {
		size, data, err := a.container.GetBlobData(digest)
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

func (a *ArtefactImpl) AddArtefact(art cpi.Artefact, platform *artdesc.Platform) (cpi.BlobAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	_, err := a.Index()
	if err != nil {
		return nil, err
	}
	return NewIndexForArtefact(a).AddArtefact(art, platform)
}

func (a *ArtefactImpl) AddLayer(blob cpi.BlobAccess, d *cpi.Descriptor) (int, error) {
	if a.IsClosed() {
		return -1, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return -1, accessio.ErrReadOnly
	}
	_, err := a.Manifest()
	if err != nil {
		return -1, err
	}
	return NewManifestForArtefact(a).AddLayer(blob, d)
}

func AdjustSize(d *artdesc.Descriptor, size int64) error {
	if size != accessio.BLOB_UNKNOWN_SIZE {
		if d.Size == accessio.BLOB_UNKNOWN_SIZE {
			d.Size = size
		} else {
			if d.Size != size {
				return errors.Newf("blob size mismatch %d != %d", size, d.Size)
			}
		}
	}
	return nil
}
