// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/core"
	"github.com/open-component-model/ocm/pkg/errors"
)

var ErrNoIndex = errors.New("manifest does not support access to subsequent artefacts")

type ArtefactImpl struct {
	artefactBase
}

var _ ArtefactAccess = (*ArtefactImpl)(nil)

func NewArtefactForProviderBlob(access ArtefactSetContainer, p ArtefactProvider, blob accessio.BlobAccess) (*ArtefactImpl, error) {
	mode := accessobj.ACC_WRITABLE
	if access.IsReadOnly() {
		mode = accessobj.ACC_READONLY
	}
	state, err := accessobj.NewBlobStateForBlob(mode, blob, NewArtefactStateHandler())
	if err != nil {
		return nil, err
	}
	a := &ArtefactImpl{
		artefactBase: artefactBase{
			container: access,
			state:     state,
			provider:  p,
		},
	}
	return a, nil
}

func NewArtefactForBlob(access ArtefactSetContainer, blob accessio.BlobAccess) (*ArtefactImpl, error) {
	mode := accessobj.ACC_WRITABLE
	if access.IsReadOnly() {
		mode = accessobj.ACC_READONLY
	}
	state, err := accessobj.NewBlobStateForBlob(mode, blob, NewArtefactStateHandler())
	if err != nil {
		return nil, err
	}
	p, err := access.NewArtefactProvider(state)
	if err != nil {
		return nil, err
	}
	a := &ArtefactImpl{
		artefactBase: artefactBase{
			container: access,
			state:     state,
			provider:  p,
		},
	}
	return a, nil
}

func NewArtefact(access ArtefactSetContainer, defs ...*artdesc.Artefact) (ArtefactAccess, error) {
	var def *artdesc.Artefact
	if len(defs) != 0 && defs[0] != nil {
		def = defs[0]
	}
	mode := accessobj.ACC_WRITABLE
	if access.IsReadOnly() {
		mode = accessobj.ACC_READONLY
	}
	state, err := accessobj.NewBlobStateForObject(mode, def, NewArtefactStateHandler())
	if err != nil {
		panic("oops: " + err.Error())
	}

	p, err := access.NewArtefactProvider(state)
	if err != nil {
		return nil, err
	}
	a := &ArtefactImpl{
		artefactBase: artefactBase{
			container: access,
			provider:  p,
			state:     state,
		},
	}
	return a, nil
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (a *ArtefactImpl) AddBlob(access BlobAccess) error {
	return a.addBlob(access)
}

func (a *ArtefactImpl) NewArtefact(art ...*artdesc.Artefact) (ArtefactAccess, error) {
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

func (a *ArtefactImpl) GetBlobDescriptor(digest digest.Digest) *Descriptor {
	d := a.GetDescriptor().GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return a.provider.GetBlobDescriptor(digest)
	// return a.container.GetBlobDescriptor(digest)
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

func (a *ArtefactImpl) GetArtefact(digest digest.Digest) (ArtefactAccess, error) {
	if !a.IsIndex() {
		return nil, ErrNoIndex
	}
	return a.getArtefact(digest)
}

func (a *ArtefactImpl) GetBlobData(digest digest.Digest) (int64, DataAccess, error) {
	return a.provider.GetBlobData(digest)
}

func (a *ArtefactImpl) GetBlob(digest digest.Digest) (BlobAccess, error) {
	d := a.GetBlobDescriptor(digest)
	if d != nil {
		size, data, err := a.provider.GetBlobData(digest)
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

func (a *ArtefactImpl) AddArtefact(art Artefact, platform *artdesc.Platform) (accessio.BlobAccess, error) {
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

func (a *ArtefactImpl) AddLayer(blob BlobAccess, d *Descriptor) (int, error) {
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
		} else if d.Size != size {
			return errors.Newf("blob size mismatch %d != %d", size, d.Size)
		}
	}
	return nil
}
