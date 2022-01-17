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

package ctf

import (
	"fmt"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/core"
)

var ErrOCIArtefatsNotSupported = errors.ErrNotSupported("oci artefacts", "plain component")

type plainComponentSpec struct{}

var _ core.RepositorySpec = &plainComponentSpec{}

func (_ plainComponentSpec) GetType() string {
	return "DummyRepo"
}

func (_ plainComponentSpec) SetType(ttype string) {
	panic("not supported")
}

func (_ plainComponentSpec) GetName() string {
	return "DummyRepo"
}

func (_ plainComponentSpec) GetVersion() string {
	return "no version"
}

func (p plainComponentSpec) Repository(context core.Context) (core.Repository, error) {
	return &plainComponent{}, nil
}

////////////////////////////////////////////////////////////////////////////////

type plainComponent struct {
	ctx core.Context
	ca  *ComponentArchive
}

var _ core.Repository = &plainComponent{}

func newPlainComponent(ca *ComponentArchive, ctx core.Context) core.Repository {
	if ctx == nil {
		ctx = ca.GetContext()
	}
	return &plainComponent{
		ctx: ctx,
		ca:  ca,
	}
}

func (_ plainComponent) LocalSupportForAccessSpec(a compdesc.AccessSpec) bool {
	return a.GetName() == accessmethods.LocalBlobType
}

func (_ plainComponent) ExistsArtefact(name string, version string) (bool, error) {
	return false, nil
}

func (_ plainComponent) LookupArtefact(name string, version string) (oci.ArtefactAccess, error) {
	return nil, ErrOCIArtefatsNotSupported
}

func (_ plainComponent) ComposeArtefact(name string, version string) (oci.ArtefactComposer, error) {
	return nil, ErrOCIArtefatsNotSupported
}

func (_ plainComponent) WriteArtefact(access oci.ArtefactAccess) (oci.ArtefactAccess, error) {
	return nil, ErrOCIArtefatsNotSupported
}

func (p *plainComponent) GetContext() core.Context {
	return p.ca.GetContext()
}

func (_ plainComponent) GetSpecification() core.RepositorySpec {
	return &plainComponentSpec{}
}

func (p *plainComponent) ExistsComponent(name string, version string) (bool, error) {
	return p.ca != nil && p.ca.GetName() == name && p.ca.GetVersion() == version, nil
}

func (p *plainComponent) LookupComponent(name string, version string) (core.ComponentAccess, error) {
	if ok, _ := p.ExistsComponent(name, version); ok {
		return p.ca, nil
	}
	return nil, errors.ErrNotFound(errors.KIND_COMPONENT, fmt.Sprintf("%s/%s", name, version))
}

func (p *plainComponent) ComposeComponent(name string, version string) (core.ComponentComposer, error) {
	if ok, _ := p.ExistsComponent(name, version); ok {
		return p.ca, nil
	}
	return nil, errors.ErrNotFound(errors.KIND_COMPONENT, fmt.Sprintf("%s/%s", name, version))
}

func (_ plainComponent) WriteComponent(access core.ComponentAccess) (core.ComponentAccess, error) {
	return nil, errors.ErrNotSupported("write component", "plain component")
}
