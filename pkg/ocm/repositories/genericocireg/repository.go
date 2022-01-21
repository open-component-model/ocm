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

package genericocireg

import (
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/core"
)

type Repository struct {
	ctx     core.Context
	ocirepo oci.Repository
}

func NewRepository(ctx core.Context, ocirepo oci.Repository) (core.Repository, error) {
	repo := &Repository{
		ctx:     ctx,
		ocirepo: ocirepo,
	}
	_ = repo
	return repo, nil
}

func (r Repository) ExistsArtefact(name string, version string) (bool, error) {
	panic("implement me")
}

func (r Repository) LookupArtefact(name string, version string) (oci.ArtefactAccess, error) {
	panic("implement me")
}

func (r Repository) ComposeArtefact(name string, version string) (oci.ArtefactComposer, error) {
	panic("implement me")
}

func (r Repository) WriteArtefact(access oci.ArtefactAccess) (oci.ArtefactAccess, error) {
	panic("implement me")
}

func (r Repository) GetContext() core.Context {
	panic("implement me")
}

func (r Repository) GetSpecification() core.RepositorySpec {
	panic("implement me")
}

func (r Repository) ExistsComponent(name string, version string) (bool, error) {
	panic("implement me")
}

func (r Repository) LookupComponent(name string, version string) (core.ComponentAccess, error) {
	panic("implement me")
}

func (r Repository) ComposeComponent(name string, version string) (core.ComponentComposer, error) {
	panic("implement me")
}

func (r Repository) WriteComponent(access core.ComponentAccess) (core.ComponentAccess, error) {
	panic("implement me")
}

func (r Repository) LocalSupportForAccessSpec(a compdesc.AccessSpec) bool {
	panic("implement me")
}
