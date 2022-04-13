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

package ociblob

import (
	"io"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/runtime"
	"github.com/opencontainers/go-digest"
)

// Type is the access type for a blob in an OCI repository.
const Type = "ociBlob"
const TypeV1 = Type + runtime.VersionSeparator + "v1"

func init() {
	cpi.RegisterAccessType(cpi.NewAccessSpecType(Type, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(TypeV1, &AccessSpec{}))
}

// New creates a new OCIBlob accessor.
func New(repository string, digest digest.Digest, mediaType string, size int64) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		Reference:           repository,
		MediaType:           mediaType,
		Digest:              digest,
		Size:                size,
	}
}

// AccessSpec describes the access for a oci registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Reference is the oci reference to the manifest
	Reference string `json:"ref"`

	// MediaType is the media type of the object this schema refers to.
	MediaType string `json:"mediaType,omitempty"`

	// Digest is the digest of the targeted content.
	Digest digest.Digest `json:"digest"`

	// Size specifies the size in bytes of the blob.
	Size int64 `json:"size"`
}

var _ cpi.AccessSpec = (*AccessSpec)(nil)

func (s AccessSpec) IsLocal(context cpi.Context) bool {
	return false
}

func (s *AccessSpec) AccessMethod(access cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	return &accessMethod{access, s}, nil
}

////////////////////////////////////////////////////////////////////////////////

type accessMethod struct {
	comp cpi.ComponentVersionAccess
	spec *AccessSpec
}

var _ cpi.AccessMethod = (*accessMethod)(nil)

func (o *accessMethod) GetKind() string {
	return Type
}

func (o *accessMethod) Get() ([]byte, error) {
	return accessio.BlobData(o.getBlob())
}

func (o *accessMethod) Reader() (io.ReadCloser, error) {
	return accessio.BlobReader(o.getBlob())
}

func (o *accessMethod) MimeType() string {
	return o.MimeType()
}

func (m *accessMethod) getBlob() (cpi.BlobAccess, error) {
	ref, err := oci.ParseRef(m.spec.Reference)
	if err != nil {
		return nil, err
	}
	if ref.Tag != nil || ref.Digest != nil {
		return nil, errors.ErrInvalid("oci repository", m.spec.Reference)
	}
	spec := ocireg.NewRepositorySpec(ref.Host)
	ocirepo, err := m.comp.GetContext().OCIContext().RepositoryForSpec(spec)
	if err != nil {
		return nil, err
	}
	ns, err := ocirepo.LookupNamespace(ref.Repository)
	if err != nil {
		return nil, err
	}
	acc, err := ns.GetBlobData(m.spec.Digest)
	if err != nil {
		return nil, err
	}
	if m.spec.Size <= 0 {
		m.spec.Size = -1
	}
	return accessio.BlobAccessForDataAccess(m.spec.Digest, m.spec.Size, m.spec.MediaType, acc), nil
}
