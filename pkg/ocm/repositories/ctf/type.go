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

package ctf

import (
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/runtime"
)

const (
	CTFComponentArchiveType   = "Component Archive"
	CTFComponentArchiveTypeV1 = CTFComponentArchiveType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(CTFComponentArchiveType, cpi.NewRepositoryType(CTFComponentArchiveType, &RepositorySpec{}, nil))
	cpi.RegisterRepositoryType(CTFComponentArchiveTypeV1, cpi.NewRepositoryType(CTFComponentArchiveTypeV1, &RepositorySpec{}, nil))
}

type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	accessio.Options            `json:",inline"`

	// FileFormat is the format of the repository file
	FilePath string `json:"filePath"`
	// AccessMode can be set to request readonly access or creation
	AccessMode accessobj.AccessMode `json:"accessMode,omitempty"`
}

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec(filePath string, opts ...accessio.Option) *RepositorySpec {
	o := accessio.Options{}
	o.ApplyOptions(opts...)
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(CTFComponentArchiveType),
		FilePath:            filePath,
		Options:             o.Default(),
	}
}

func (a *RepositorySpec) GetType() string {
	return CTFComponentArchiveType
}
func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return NewRepository(ctx, a)
}
