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
	"fmt"

	"github.com/gardener/ocm/pkg/ocm/common"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
// to OCI Image References.
type ComponentNameMapping string

const (
	OCIRegistryRepositoryType   = "OCIRegistry"
	OCIRegistryRepositoryTypeV1 = OCIRegistryRepositoryType + "/v1"

	OCIRegistryURLPathMapping ComponentNameMapping = "urlPath"
	OCIRegistryDigestMapping  ComponentNameMapping = "sha256-digest"
)

func init() {
	core.RegisterRepositoryType(OCIRegistryRepositoryType, common.NewRepositoryType(OCIRegistryRepositoryType, &OCIRegistryRepositorySpec{}))
	core.RegisterRepositoryType(OCIRegistryRepositoryTypeV1, common.NewRepositoryType(OCIRegistryRepositoryTypeV1, &OCIRegistryRepositorySpec{}))
}

// OCIRegistryRepositorySpec describes a component repository backed by a oci registry.
type OCIRegistryRepositorySpec struct {
	runtime.ObjectTypeVersion `json:",inline"`
	// BaseURL is the base url of the repository to resolve components.
	BaseURL string `json:"baseUrl"`
	// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
	// to OCI Image References.
	ComponentNameMapping ComponentNameMapping `json:"componentNameMapping"`
}

// NewOCIRegistryRepositorySpec creates a new OCIRegistryRepositorySpec
func NewOCIRegistryRepositorySpec(baseURL string, mapping ComponentNameMapping) *OCIRegistryRepositorySpec {
	if len(mapping) == 0 {
		mapping = OCIRegistryURLPathMapping
	}
	return &OCIRegistryRepositorySpec{
		ObjectTypeVersion:    runtime.NewObjectTypeVersion(OCIRegistryRepositoryType),
		BaseURL:              baseURL,
		ComponentNameMapping: mapping,
	}
}

func (a *OCIRegistryRepositorySpec) GetType() string {
	return OCIRegistryRepositoryType
}
func (a *OCIRegistryRepositorySpec) Repository() (core.Repository, error) {
	return nil, fmt.Errorf("NOT IMPLEMENTED") // TODO
}
