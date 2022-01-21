// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ocireg

import (
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	area "github.com/gardener/ocm/pkg/ocm/core"
	areautils "github.com/gardener/ocm/pkg/ocm/ocmutils"
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
	area.RegisterRepositoryType(OCIRegistryRepositoryType, areautils.NewRepositoryType(OCIRegistryRepositoryType, &OCIRegistryRepositorySpec{}, localAccessChecker))
	area.RegisterRepositoryType(OCIRegistryRepositoryTypeV1, areautils.NewRepositoryType(OCIRegistryRepositoryTypeV1, &OCIRegistryRepositorySpec{}, localAccessChecker))
}

// ComponentRepositoryMeta describes config special for a mapping of
// a component repository to an oci registry
type ComponentRepositoryMeta struct {
	// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
	// to OCI Image References.
	ComponentNameMapping ComponentNameMapping `json:"componentNameMapping"`
}

// OCIRegistryRepositorySpec describes a component repository backed by a oci registry.
type OCIRegistryRepositorySpec struct {
	ocireg.OCIRegistryRepositorySpec `json:",inline"`
	ComponentRepositoryMeta          `json:",inline"`
}

func NewComponentRepositoryMeta(mapping ComponentNameMapping) *ComponentRepositoryMeta {
	return &ComponentRepositoryMeta{
		ComponentNameMapping: mapping,
	}
}

// NewOCIRegistryRepositorySpec creates a new OCIRegistryRepositorySpec
func NewOCIRegistryRepositorySpec(baseURL string, mapping ComponentNameMapping) *OCIRegistryRepositorySpec {
	if len(mapping) == 0 {
		mapping = OCIRegistryURLPathMapping
	}
	return &OCIRegistryRepositorySpec{
		OCIRegistryRepositorySpec: *ocireg.NewOCIRegistryRepositorySpec(baseURL),
		ComponentRepositoryMeta:   *NewComponentRepositoryMeta(mapping),
	}
}

func (a *OCIRegistryRepositorySpec) GetType() string {
	return OCIRegistryRepositoryType
}
func (a *OCIRegistryRepositorySpec) Repository(ctx area.Context) (area.Repository, error) {
	return nil, errors.ErrNotImplemented() // TODO
}

func localAccessChecker(ctx area.Context, a compdesc.AccessSpec) bool {
	name := a.GetName()
	return name == accessmethods.LocalBlobType
}
