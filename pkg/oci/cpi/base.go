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
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/opencontainers/go-digest"
)

type artefactBase struct {
	lock   sync.RWMutex
	access ArtefactSetContainer
	state  accessobj.State
}

func (a *artefactBase) IsClosed() bool {
	return a.access.IsClosed()
}

func (a *artefactBase) IsReadOnly() bool {
	return a.access.IsReadOnly()
}

func (a *artefactBase) IsIndex() bool {
	d := a.state.GetState().(*artdesc.Artefact)
	return d.IsIndex()
}

func (a *artefactBase) IsManifest() bool {
	d := a.state.GetState().(*artdesc.Artefact)
	return d.IsManifest()
}

func (a *artefactBase) blob() (accessio.BlobAccess, error) {
	return a.state.GetBlob()
}

func (a *artefactBase) addBlob(access BlobAccess) error {
	return a.access.AddBlob(access)
}

func (a *artefactBase) getArtefact(digest digest.Digest) (ArtefactAccess, error) {
	return a.access.GetArtefact(digest.String())
}

func (a *artefactBase) newArtefact(art ...*artdesc.Artefact) (ArtefactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return NewArtefact(a.access, art...), nil
}
