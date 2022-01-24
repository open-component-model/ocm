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

package memory

import (
	"github.com/gardener/ocm/pkg/credentials/cpi"
	"github.com/gardener/ocm/pkg/runtime"
)

const (
	MemoryRepositoryType   = "Memory"
	MemoryRepositoryTypeV1 = MemoryRepositoryType + "/v1"
)

func init() {
	cpi.RegisterRepositoryType(MemoryRepositoryType, cpi.NewRepositoryType(MemoryRepositoryType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(MemoryRepositoryTypeV1, cpi.NewRepositoryType(MemoryRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes a memory based repository interface.
type RepositorySpec struct {
	runtime.ObjectTypeVersion `json:",inline"`
	RepositoryName            string `json:"repoName"`
}

// NewRepositorySpec creates a new memory RepositorySpec
func NewRepositorySpec(name string) *RepositorySpec {
	return &RepositorySpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(MemoryRepositoryType),
		RepositoryName:    name,
	}
}

func (a *RepositorySpec) GetType() string {
	return MemoryRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	repos := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
	return repos.GetRepository(a.RepositoryName), nil
}
