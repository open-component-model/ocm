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

package accessmethods

import (
	"io"

	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/runtime"
	"github.com/opencontainers/go-digest"
)

// OCIBlobType is the access type for a blob in an OCI repository.
const OCIBlobType = "ociBlob"
const OCIBlobTypeV1 = OCIBlobType + runtime.VersionSeparator + "v1"

func init() {
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(LocalBlobType, LocalBlobV1))
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(LocalBlobTypeV1, LocalBlobV1))
}

// NewOCIBlobAccessSpec creates a new OCIBlob accessor.
func NewOCIBlobAccessSpec(repository string, digest digest.Digest, mediaType string, size int64) *OCIBlobAccessSpec {
	return &OCIBlobAccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(OCIBlobType),
		Reference:           repository,
		MediaType:           mediaType,
		Digest:              digest,
		Size:                size,
	}
}

// OCIBlobAccessSpec describes the access for a oci registry.
type OCIBlobAccessSpec struct {
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

var _ cpi.AccessSpec = (*OCIBlobAccessSpec)(nil)

func (s OCIBlobAccessSpec) IsLocal(context core.Context) bool {
	return false
}

func (s *OCIBlobAccessSpec) AccessMethod(access core.ComponentVersionAccess) (core.AccessMethod, error) {
	return &ociBlobAccessMethod{s}, nil
}

////////////////////////////////////////////////////////////////////////////////

type ociBlobAccessMethod struct {
	spec *OCIBlobAccessSpec
}

var _ cpi.AccessMethod = (*ociBlobAccessMethod)(nil)

func (o *ociBlobAccessMethod) GetKind() string {
	return OCIBlobType
}

func (o *ociBlobAccessMethod) Get() ([]byte, error) {
	panic("implement me")
}

func (o *ociBlobAccessMethod) Reader() (io.ReadCloser, error) {
	panic("implement me")
}

func (o *ociBlobAccessMethod) MimeType() string {
	return o.MimeType()
}
