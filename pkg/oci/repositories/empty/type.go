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
	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/runtime"
)

const (
	EmptyRepositoryType   = "Empty"
	EmptyRepositoryTypeV1 = EmptyRepositoryType + "/v1"
)

const ATTR_REPOS = "github.com/gardener/ocm/pkg/oci/repositories/empty"

func init() {
	cpi.RegisterRepositoryType(EmptyRepositoryType, cpi.NewRepositoryType(EmptyRepositoryType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(EmptyRepositoryTypeV1, cpi.NewRepositoryType(EmptyRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes an OCI registry interface backed by an oci registry.
type RepositorySpec struct {
	runtime.ObjectTypeVersion `json:",inline"`
}

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec() *RepositorySpec {
	return &RepositorySpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(EmptyRepositoryType),
	}
}

func (a *RepositorySpec) GetType() string {
	return EmptyRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepository).(*Repository), nil
}
