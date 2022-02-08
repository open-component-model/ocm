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

package ocireg

import (
	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/gardener/ocm/pkg/oci/cpi"
)

type Repository struct {
	ctx  cpi.Context
	spec *RepositorySpec
}

var _ cpi.Repository = &Repository{}

func NewRepository(ctx cpi.Context, spec *RepositorySpec, creds credentials.Credentials) (*Repository, error) {
	return &Repository{
		ctx:  ctx,
		spec: spec,
	}, nil
}

func (r *Repository) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *Repository) SupportsDistributionSpec() bool {
	return true
}

func (r *Repository) ExistsArtefact(name string, version string) (bool, error) {
	panic("implement me")
}

func (r *Repository) LookupArtefact(name string, version string) (core.ArtefactAccess, error) {
	panic("implement me")
}

func (r *Repository) LookupNamespace(name string) (core.NamespaceAccess, error) {
	panic("implement me")
}

func (r *Repository) Close() error {
	panic("implement me")
}
