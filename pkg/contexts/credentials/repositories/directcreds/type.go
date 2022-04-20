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

package directcreds

import (
	"github.com/open-component-model/ocm/pkg/common"
	core2 "github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	cpi2 "github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	DirectCredentialsRepositoryType   = cpi2.DirectCredentialsType
	DirectCredentialsRepositoryTypeV1 = DirectCredentialsRepositoryType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi2.RegisterRepositoryType(DirectCredentialsRepositoryType, cpi2.NewRepositoryType(DirectCredentialsRepositoryType, &RepositorySpec{}))
	cpi2.RegisterRepositoryType(DirectCredentialsRepositoryTypeV1, cpi2.NewRepositoryType(DirectCredentialsRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes a repository interface for single direct credentials.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	Properties                  common.Properties `json:"properties"`
}

var _ cpi2.RepositorySpec = &RepositorySpec{}
var _ cpi2.CredentialsSpec = &RepositorySpec{}

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec(credentials common.Properties) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(DirectCredentialsRepositoryType),
		Properties:          credentials,
	}
}

func (a *RepositorySpec) GetType() string {
	return DirectCredentialsRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi2.Context, creds cpi2.Credentials) (cpi2.Repository, error) {
	return NewRepository(cpi2.NewCredentials(a.Properties)), nil
}

func (a *RepositorySpec) Credentials(context core2.Context, source ...core2.CredentialsSource) (core2.Credentials, error) {
	return cpi2.NewCredentials(a.Properties), nil
}

func (a *RepositorySpec) GetCredentialsName() string {
	return ""
}

func (a *RepositorySpec) GetRepositorySpec(context core2.Context) core2.RepositorySpec {
	return a
}
