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

package comparch

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	cpi2 "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	CTFComponentArchiveType   = "ComponentArchive"
	CTFComponentArchiveTypeV1 = CTFComponentArchiveType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi2.RegisterRepositoryType(CTFComponentArchiveType, cpi2.NewRepositoryType(CTFComponentArchiveType, &RepositorySpec{}, nil))
	cpi2.RegisterRepositoryType(CTFComponentArchiveTypeV1, cpi2.NewRepositoryType(CTFComponentArchiveTypeV1, &RepositorySpec{}, nil))
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
func NewRepositorySpec(acc accessobj.AccessMode, filePath string, opts ...accessio.Option) *RepositorySpec {
	o := accessio.AccessOptions(opts...)
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(CTFComponentArchiveType),
		FilePath:            filePath,
		Options:             o,
		AccessMode:          acc,
	}
}

func (a *RepositorySpec) GetType() string {
	return CTFComponentArchiveType
}
func (a *RepositorySpec) Repository(ctx cpi2.Context, creds credentials.Credentials) (cpi2.Repository, error) {
	return NewRepository(ctx, a)
}
func (a *RepositorySpec) AsUniformSpec(cpi2.Context) cpi2.UniformRepositorySpec {
	opts := a.Options.Default()
	p, err := vfs.Canonical(opts.PathFileSystem, a.FilePath, false)
	if err != nil {
		return cpi2.UniformRepositorySpec{Type: a.GetKind(), SubPath: a.FilePath}
	}
	return cpi2.UniformRepositorySpec{Type: a.GetKind(), SubPath: p}
}
