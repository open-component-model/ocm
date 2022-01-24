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

package aliases

import (
	"github.com/gardener/ocm/pkg/credentials/cpi"
	"github.com/gardener/ocm/pkg/runtime"
)

const (
	AliasRepositoryType   = cpi.AliasRepositoryType
	AliasRepositoryTypeV1 = AliasRepositoryType + "/v1"
)

func init() {
	cpi.RegisterRepositoryType(AliasRepositoryType, cpi.NewAliasRegistry(cpi.NewRepositoryType(AliasRepositoryType, &RepositorySpec{}), setAlias))
	cpi.RegisterRepositoryType(AliasRepositoryTypeV1, cpi.NewRepositoryType(AliasRepositoryTypeV1, &RepositorySpec{}))
}

func setAlias(ctx cpi.Context, name string, spec cpi.RepositorySpec, creds cpi.CredentialsSource) error {
	repos := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
	repos.Set(name, spec, creds)
	return nil
}

// RepositorySpec describes a memory based repository interface.
type RepositorySpec struct {
	runtime.ObjectTypeVersion `json:",inline"`
	Alias                     string `json:"alias"`
}

// NewRepositorySpec creates a new memory RepositorySpec
func NewRepositorySpec(name string) *RepositorySpec {
	return &RepositorySpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(AliasRepositoryType),
		Alias:             name,
	}
}

func (a *RepositorySpec) GetType() string {
	return AliasRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	repos := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
	alias := repos.GetRepository(a.Alias)
	if alias == nil {
		return nil, cpi.ErrUnknownRepository(AliasRepositoryType, a.Alias)
	}
	return alias.GetRepository(ctx, creds)
}
