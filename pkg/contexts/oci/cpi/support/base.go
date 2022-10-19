// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package support

import (
	"sync"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type artefactBase struct {
	lock      sync.RWMutex
	view      ArtefactSetContainer
	container ArtefactSetContainerImpl
	state     accessobj.State
}

func newArtefactBase(view ArtefactSetContainer, container ArtefactSetContainerImpl, state accessobj.State) *artefactBase {
	return &artefactBase{
		view:      view,
		container: container,
		state:     state,
	}
}

func (a *artefactBase) IsClosed() bool {
	return a.view.IsClosed()
}

func (a *artefactBase) IsReadOnly() bool {
	return a.container.IsReadOnly()
}

func (a *artefactBase) IsIndex() bool {
	d := a.state.GetState().(*artdesc.Artefact)
	return d.IsIndex()
}

func (a *artefactBase) IsManifest() bool {
	d := a.state.GetState().(*artdesc.Artefact)
	return d.IsManifest()
}

func (a *artefactBase) blob() (cpi.BlobAccess, error) {
	return a.state.GetBlob()
}

func (a *artefactBase) addBlob(access cpi.BlobAccess) error {
	return a.container.AddBlob(access)
}

func (a *artefactBase) newArtefact(art ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return NewArtefact(a.container, art...)
}

func (a *artefactBase) Blob() (accessio.BlobAccess, error) {
	d := a.state.GetState().(artdesc.BlobDescriptorSource)
	if !d.IsValid() {
		return nil, errors.ErrUnknown("artefact type")
	}
	blob, err := a.blob()
	if err != nil {
		return nil, err
	}
	return accessio.BlobWithMimeType(d.MimeType(), blob), nil
}

func (a *artefactBase) Digest() digest.Digest {
	d := a.state.GetState().(artdesc.BlobDescriptorSource)
	if !d.IsValid() {
		return ""
	}
	blob, err := a.blob()
	if err != nil {
		return ""
	}
	return blob.Digest()
}
