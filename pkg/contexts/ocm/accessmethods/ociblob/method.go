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
	"sync"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
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
	return &accessMethod{comp: access, spec: s}, nil
}

////////////////////////////////////////////////////////////////////////////////

// TODO add cache

type accessMethod struct {
	lock sync.Mutex
	blob accessio.BlobAccess
	comp cpi.ComponentVersionAccess
	spec *AccessSpec
}

var _ cpi.AccessMethod = (*accessMethod)(nil)

func (m *accessMethod) GetKind() string {
	return Type
}

func (m *accessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		m.blob.Close()
		m.blob = nil
	}
	return nil
}

func (m *accessMethod) Get() ([]byte, error) {
	return accessio.BlobData(m.getBlob())
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	return accessio.BlobReader(m.getBlob())
}

func (m *accessMethod) MimeType() string {
	return m.MimeType()
}

func (m *accessMethod) getBlob() (cpi.BlobAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.blob != nil {
		return m.blob, nil
	}
	ref, err := oci.ParseRef(m.spec.Reference)
	if err != nil {
		return nil, err
	}
	if ref.Tag != nil || ref.Digest != nil {
		return nil, errors.ErrInvalid("oci repository", m.spec.Reference)
	}
	ocictx := m.comp.GetContext().OCIContext()
	spec := ocictx.GetAlias(ref.Host)
	if spec == nil {
		spec = ocireg.NewRepositorySpec(ref.Host)
	}
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
	m.blob = accessio.BlobAccessForDataAccess(m.spec.Digest, m.spec.Size, m.spec.MediaType, acc)
	return m.blob, nil
}
