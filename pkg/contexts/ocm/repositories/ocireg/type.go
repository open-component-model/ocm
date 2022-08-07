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
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
)

// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
// to OCI Image References.
type ComponentNameMapping = genericocireg.ComponentNameMapping

const (
	Type   = genericocireg.Type
	TypeV1 = genericocireg.TypeV1

	OCIRegistryURLPathMapping ComponentNameMapping = "urlPath"
	OCIRegistryDigestMapping  ComponentNameMapping = "sha256-digest"
)

// ComponentRepositoryMeta describes config special for a mapping of
// a component repository to an oci registry
type ComponentRepositoryMeta = genericocireg.ComponentRepositoryMeta

// RepositorySpec describes a component repository backed by a oci registry.
type RepositorySpec = genericocireg.RepositorySpec

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec(baseURL string, meta *ComponentRepositoryMeta) *RepositorySpec {
	return genericocireg.NewRepositorySpec(ocireg.NewRepositorySpec(baseURL), meta)
}

func NewComponentRepositoryMeta(subPath string, mapping ComponentNameMapping) *ComponentRepositoryMeta {
	return genericocireg.NewComponentRepositoryMeta(subPath, mapping)
}
