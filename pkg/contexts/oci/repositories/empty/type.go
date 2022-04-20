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

package empty

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	cpi2 "github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	EmptyRepositoryType   = "Empty"
	EmptyRepositoryTypeV1 = EmptyRepositoryType + runtime.VersionSeparator + "v1"
)

const ATTR_REPOS = "github.com/open-component-model/ocm/pkg/oci/repositories/empty"

func init() {
	cpi2.RegisterRepositoryType(EmptyRepositoryType, cpi2.NewRepositoryType(EmptyRepositoryType, &RepositorySpec{}))
	cpi2.RegisterRepositoryType(EmptyRepositoryTypeV1, cpi2.NewRepositoryType(EmptyRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes an OCI registry interface backed by an oci registry.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
}

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec() *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(EmptyRepositoryType),
	}
}

func (a *RepositorySpec) GetType() string {
	return EmptyRepositoryType
}

func (a *RepositorySpec) Name() string {
	return EmptyRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi2.Context, creds credentials.Credentials) (cpi2.Repository, error) {
	return ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepository).(*Repository), nil
}
