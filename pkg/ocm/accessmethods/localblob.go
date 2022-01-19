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
	"encoding/json"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/core/accesstypes"
	"github.com/gardener/ocm/pkg/runtime"
)

// LocalBlobType is the access type of a blob local to a component.
const LocalBlobType = "localBlob"
const LocalBlobTypeV1 = LocalBlobType + "/v1"

func init() {
	core.RegisterAccessType(accesstypes.NewConvertedType(LocalBlobType, LocalBlobV1))
	core.RegisterAccessType(accesstypes.NewConvertedType(LocalBlobTypeV1, LocalBlobV1))
}

// NewLocalBlobAccessSpecV1 creates a new localFilesystemBlob accessor.
func NewLocalBlobAccessSpecV1(path string, mediaType string) *LocalBlobAccessSpec {
	return &LocalBlobAccessSpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(LocalBlobType),
		Filename:          path,
		MediaType:         mediaType,
	}
}

// LocalBlobAccessSpec describes the access for a blob on the filesystem.
type LocalBlobAccessSpec struct {
	runtime.ObjectTypeVersion
	// Filename is the name of the blob in the local filesystem.
	// The blob is expected to be at <fs-root>/blobs/<name>
	Filename string
	// MediaType is the media type of the object this filename refers to.
	MediaType string
}

var _ json.Marshaler = &LocalBlobAccessSpec{}

func (s *LocalBlobAccessSpec) MarshalJSON() ([]byte, error) {
	return accesstypes.MarshalConvertedAccessSpec(s)
}

func (a *LocalBlobAccessSpec) ValidFor(repo core.Repository) bool {
	return repo.LocalSupportForAccessSpec(a)
}

func (a *LocalBlobAccessSpec) AccessMethod(c core.ComponentAccess) (core.AccessMethod, error) {
	if a.ValidFor(c.GetRepository()) {
		return c.AccessMethod(a)
	}
	return nil, errors.ErrNotImplemented(errors.KIND_ACCESSMETHOD, LocalBlobType, c.GetRepository().GetSpecification().GetName())
}

////////////////////////////////////////////////////////////////////////////////

type LocalBlobAccessSpecV1 struct {
	runtime.ObjectTypeVersion `json:",inline"`
	// Filename is the name of the blob in the local filesystem.
	// The blob is expected to be at <fs-root>/blobs/<name>
	Filename string `json:"filename"`
	// MediaType is the media type of the object this filename refers to.
	MediaType string `json:"mediaType,omitempty"`
}

type localblobConverterV1 struct{}

var LocalBlobV1 = accesstypes.NewAccessSpecVersion(&LocalBlobAccessSpecV1{}, localblobConverterV1{})

func (_ localblobConverterV1) ConvertFrom(object core.AccessSpec) (runtime.TypedObject, error) {
	in := object.(*LocalBlobAccessSpec)
	return &LocalBlobAccessSpecV1{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(in.Type),
		Filename:          in.Filename,
		MediaType:         in.MediaType,
	}, nil
}

func (_ localblobConverterV1) ConvertTo(object interface{}) (core.AccessSpec, error) {
	in := object.(*LocalBlobAccessSpecV1)
	return &LocalBlobAccessSpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(in.Type),
		Filename:          in.Filename,
		MediaType:         in.MediaType,
	}, nil
}
