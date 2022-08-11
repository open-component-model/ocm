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

package directcreds

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	Type   = "Credentials"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(Type, cpi.NewRepositoryType(Type, &RepositorySpec{}))
	cpi.RegisterRepositoryType(TypeV1, cpi.NewRepositoryType(TypeV1, &RepositorySpec{}))
}

// RepositorySpec describes a repository interface for single direct credentials.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	Properties                  common.Properties `json:"properties"`
}

var _ cpi.RepositorySpec = &RepositorySpec{}
var _ cpi.CredentialsSpec = &RepositorySpec{}

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec(credentials common.Properties) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		Properties:          credentials,
	}
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	return NewRepository(cpi.NewCredentials(a.Properties)), nil
}

func (a *RepositorySpec) Credentials(context cpi.Context, source ...cpi.CredentialsSource) (cpi.Credentials, error) {
	return cpi.NewCredentials(a.Properties), nil
}

func (a *RepositorySpec) GetCredentialsName() string {
	return ""
}

func (a *RepositorySpec) GetRepositorySpec(context cpi.Context) cpi.RepositorySpec {
	return a
}
