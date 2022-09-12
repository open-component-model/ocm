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
