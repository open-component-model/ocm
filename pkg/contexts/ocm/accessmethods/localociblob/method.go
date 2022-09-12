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

package localociblob

import (
	"fmt"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type for a component version local blob in an OCI repository.
const Type = "localOciBlob"
const TypeV1 = Type + runtime.VersionSeparator + "v1"

func init() {
	cpi.RegisterAccessType(cpi.NewAccessSpecType(Type, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(TypeV1, &AccessSpec{}))
}

// New creates a new LocalOCIBlob accessor.
// Deprecated: Use LocalBlob.
func New(digest digest.Digest) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		Digest:              digest,
	}
}

// AccessSpec describes the access for a oci registry.
// Deprecated: Use LocalBlob.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Digest is the digest of the targeted content.
	Digest digest.Digest `json:"digest"`
}

var _ cpi.AccessSpec = (*AccessSpec)(nil)

func (a *AccessSpec) Describe(ctx cpi.Context) string {
	return fmt.Sprintf("Local OCI blob %s", a.Digest)
}

func (s AccessSpec) IsLocal(context cpi.Context) bool {
	return true
}

func (s *AccessSpec) GetMimeType() string {
	return ""
}

func (s *AccessSpec) AccessMethod(c cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	return c.AccessMethod(s)
}
