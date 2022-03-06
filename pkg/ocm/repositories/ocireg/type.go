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
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/cpi"
)

// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
// to OCI Image References.
type ComponentNameMapping string

const (
	OCIRegistryRepositoryType   = ocireg.OCIRegistryRepositoryType
	OCIRegistryRepositoryTypeV1 = ocireg.OCIRegistryRepositoryTypeV1

	OCIRegistryURLPathMapping ComponentNameMapping = "urlPath"
	OCIRegistryDigestMapping  ComponentNameMapping = "sha256-digest"
)

func init() {
	cpi.RegisterRepositoryType(OCIRegistryRepositoryType, cpi.NewRepositoryType(OCIRegistryRepositoryType, &RepositorySpec{}, localAccessChecker))
	cpi.RegisterRepositoryType(OCIRegistryRepositoryTypeV1, cpi.NewRepositoryType(OCIRegistryRepositoryTypeV1, &RepositorySpec{}, localAccessChecker))
}

// ComponentRepositoryMeta describes config special for a mapping of
// a component repository to an oci registry
type ComponentRepositoryMeta struct {
	// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
	// to OCI Image References.
	ComponentNameMapping ComponentNameMapping `json:"componentNameMapping,omitempty"`
	SubPath              string               `json:"subPath,omitempty"`
}

// RepositorySpec describes a component repository backed by a oci registry.
type RepositorySpec struct {
	ocireg.RepositorySpec   `json:",inline"`
	ComponentRepositoryMeta `json:",inline"`
}

func NewComponentRepositoryMeta(subPath string, mapping ComponentNameMapping) *ComponentRepositoryMeta {
	return &ComponentRepositoryMeta{
		ComponentNameMapping: mapping,
		SubPath:              subPath,
	}
}

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec(baseURL string, mapping ComponentNameMapping) *RepositorySpec {
	if len(mapping) == 0 {
		mapping = OCIRegistryURLPathMapping
	}
	return &RepositorySpec{
		RepositorySpec:          *ocireg.NewRepositorySpec(baseURL),
		ComponentRepositoryMeta: *NewComponentRepositoryMeta("", mapping),
	}
}

func (a *RepositorySpec) GetType() string {
	return OCIRegistryRepositoryType
}
func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return nil, errors.ErrNotImplemented() // TODO
}

func localAccessChecker(ctx cpi.Context, a compdesc.AccessSpec) bool {
	name := a.GetKind()
	return name == accessmethods.LocalBlobType
}
