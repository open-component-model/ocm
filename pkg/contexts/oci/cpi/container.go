// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
)

// ArtefactProvider manages the technical access to a dedicated artefact.
type ArtefactProvider interface {
	IsClosed() bool
	IsReadOnly() bool
	GetBlobDescriptor(digest digest.Digest) *Descriptor
	Close() error

	BlobSource
	BlobSink

	// GetArtefact is used to access nested artefacts (only)
	GetArtefact(digest digest.Digest) (ArtefactAccess, error)
	// AddArtefact is used to add nested artefacts (only)
	AddArtefact(art Artefact) (access accessio.BlobAccess, err error)
}

type NopCloserArtefactProvider struct {
	ArtefactSetContainer
}

var _ ArtefactProvider = (*NopCloserArtefactProvider)(nil)

func (p *NopCloserArtefactProvider) Close() error {
	return nil
}

func (p *NopCloserArtefactProvider) AddArtefact(art Artefact) (access accessio.BlobAccess, err error) {
	return p.ArtefactSetContainer.AddArtefact(art)
}

func (p *NopCloserArtefactProvider) GetArtefact(digest digest.Digest) (ArtefactAccess, error) {
	return p.ArtefactSetContainer.GetArtefact("@" + digest.String())
}

func NewNopCloserArtefactProvider(p ArtefactSetContainer) ArtefactProvider {
	return &NopCloserArtefactProvider{
		p,
	}
}

////////////////////////////////////////////////////////////////////////////////

// ArtefactSetContainer is the interface used by subsequent access objects
// to access the base implementation.
type ArtefactSetContainer interface {
	IsReadOnly() bool
	IsClosed() bool

	Close() error

	GetBlobDescriptor(digest digest.Digest) *Descriptor
	GetBlobData(digest digest.Digest) (int64, DataAccess, error)
	AddBlob(blob BlobAccess) error

	GetArtefact(vers string) (ArtefactAccess, error)
	AddArtefact(artefact Artefact, tags ...string) (access accessio.BlobAccess, err error)

	NewArtefactProvider(state accessobj.State) (ArtefactProvider, error)
}

////////////////////////////////////////////////////////////////////////////////

type artefactSetContainerImpl struct {
	refs accessio.ReferencableCloser
	ArtefactSetContainer
}

func NewArtefactSetContainer(c ArtefactSetContainer) ArtefactSetContainer {
	i := &artefactSetContainerImpl{
		refs:                 accessio.NewRefCloser(c, true),
		ArtefactSetContainer: c,
	}
	v, _ := i.View()
	return v
}

func (i *artefactSetContainerImpl) View() (ArtefactSetContainer, error) {
	v, err := i.refs.View()
	if err != nil {
		return nil, err
	}
	return &artefactSetContainerView{
		view:                 v,
		ArtefactSetContainer: i.ArtefactSetContainer,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type artefactSetContainerView struct {
	view accessio.CloserView
	ArtefactSetContainer
}

func (v *artefactSetContainerView) IsClosed() bool {
	return v.view.IsClosed()
}

func (v *artefactSetContainerView) Close() error {
	return v.view.Close()
}
