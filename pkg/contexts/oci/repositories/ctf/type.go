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

package ctf

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	Type   = cpi.CommonTransportFormat
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(Type, cpi.NewRepositoryType(Type, &RepositorySpec{}))
	cpi.RegisterRepositoryType(TypeV1, cpi.NewRepositoryType(TypeV1, &RepositorySpec{}))
}

// RepositorySpec describes an OCI registry interface backed by an oci registry.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	accessio.Options            `json:",inline"`

	// FileFormat is the format of the repository file
	FilePath string `json:"filePath"`
	// AccessMode can be set to request readonly access or creation
	AccessMode accessobj.AccessMode `json:"accessMode,omitempty"`
}

var _ cpi.RepositorySpec = (*RepositorySpec)(nil)

var _ cpi.IntermediateRepositorySpecAspect = (*RepositorySpec)(nil)

// NewRepositorySpec creates a new RepositorySpec
func NewRepositorySpec(mode accessobj.AccessMode, filePath string, opts ...accessio.Option) *RepositorySpec {
	o := accessio.AccessOptions(opts...)
	if o.FileFormat == nil {
		for _, v := range SupportedFormats() {
			if strings.HasSuffix(filePath, "."+v.String()) {
				o.FileFormat = &v
				break
			}
		}
	}
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		FilePath:            filePath,
		Options:             o.Default(),
		AccessMode:          mode,
	}
}

func (a *RepositorySpec) IsIntermediate() bool {
	return true
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (s *RepositorySpec) Name() string {
	return s.FilePath
}
func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return Open(ctx, a.AccessMode, a.FilePath, 0700, a.Options)
}
