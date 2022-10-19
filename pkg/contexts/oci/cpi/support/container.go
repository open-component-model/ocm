// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package support

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
)

// ArtefactSetContainer is the interface used by subsequent access objects
// to access the base implementation.
type ArtefactSetContainer interface {
	IsReadOnly() bool
	IsClosed() bool

	Close() error

	GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor
	GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error)
	AddBlob(blob cpi.BlobAccess) error

	GetArtefact(vers string) (cpi.ArtefactAccess, error)
	AddArtefact(artefact cpi.Artefact, tags ...string) (access accessio.BlobAccess, err error)
}

////////////////////////////////////////////////////////////////////////////////

// ArtefactSetContainerInt is the implementation interface for a provider.
type ArtefactSetContainerInt ArtefactSetContainer

type artefactSetContainerImpl struct {
	refs accessio.ReferencableCloser
	ArtefactSetContainerInt
}

type ArtefactSetContainerImpl interface {
	ArtefactSetContainer
	View(main ...bool) (ArtefactSetContainer, error)
}

func NewArtefactSetContainer(c ArtefactSetContainerInt) (ArtefactSetContainer, ArtefactSetContainerImpl) {
	i := &artefactSetContainerImpl{
		refs:                    accessio.NewRefCloser(c, true),
		ArtefactSetContainerInt: c,
	}
	v, _ := i.View(true)
	return v, i
}

func (i *artefactSetContainerImpl) View(main ...bool) (ArtefactSetContainer, error) {
	v, err := i.refs.View(main...)
	if err != nil {
		return nil, err
	}
	return &artefactSetContainerView{
		view:                     v,
		ArtefactSetContainerImpl: i,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type artefactSetContainerView struct {
	view accessio.CloserView
	ArtefactSetContainerImpl
}

func (v *artefactSetContainerView) IsClosed() bool {
	return v.view.IsClosed()
}

func (v *artefactSetContainerView) Close() error {
	return v.view.Close()
}
