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

package localblob

import (
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type of a blob local to a component.
const Type = "localBlob"
const TypeV1 = Type + runtime.VersionSeparator + "v1"

func init() {
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(Type, LocalBlobV1))
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(TypeV1, LocalBlobV1))
}

// New creates a new localFilesystemBlob accessor.
func New(local, hint string, mediaType string, global cpi.AccessSpec) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		LocalReference:      local,
		ReferenceName:       hint,
		MediaType:           mediaType,
		GlobalAccess:        core.NewAccessSpecRef(global),
	}
}

// AccessSpec describes the access for a blob on the filesystem.
type AccessSpec struct {
	runtime.ObjectVersionedType
	// LocalReference is the repository local identity of the blob.
	// it is used by the repository implementation to get access
	// to the blob and if therefore specific to a dedicated repository type.
	LocalReference string `json:"localReference"`
	// MediaType is the media type of the object represented by the blob
	MediaType string `json:"mediaType"`

	// GlobalAccess is an optional field describing a possibility
	// for a global access. If given, it MUST describe a global access method.
	GlobalAccess *core.AccessSpecRef `json:"globalAccess,omitempty"`
	// ReferenceName is an optional static name the object should be
	// use in a local repository context. It is use by a repository
	// to optionally determine a globally referencable access according
	// to the OCI distribution spec. The result will be stored
	// by the repository in the field ImageReference.
	// The value is typically an OCI repository name optionally
	// followed by a colon ':' and a tag
	ReferenceName string `json:"referenceName,omitempty"`
}

var _ json.Marshaler = (*AccessSpec)(nil)
var _ cpi.HintProvider = (*AccessSpec)(nil)
var _ cpi.AccessSpec = (*AccessSpec)(nil)

func (a AccessSpec) MarshalJSON() ([]byte, error) {
	return cpi.MarshalConvertedAccessSpec(cpi.DefaultContext(), &a)
}

func (a *AccessSpec) IsLocal(cpi.Context) bool {
	return true
}

func (a *AccessSpec) GetMimeType() string {
	return a.MediaType
}

func (a *AccessSpec) GetReferenceName() string {
	return a.ReferenceName
}

func (a *AccessSpec) AccessMethod(c cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	return c.AccessMethod(a)
}

////////////////////////////////////////////////////////////////////////////////

type AccessSpecV1 struct {
	runtime.ObjectVersionedType `json:",inline"`
	// LocalReference is the repository local identity of the blob.
	// it is used by the repository implementation to get access
	// to the blob and if therefore specific to a dedicated repository type.
	LocalReference string `json:"localReference"`
	// MediaType is the media type of the object represented by the blob
	MediaType string `json:"mediaType"`

	// GlobalAccess is an optional field describing a possibility
	// for a global access. If given, it MUST describe a global access method.
	GlobalAccess *core.AccessSpecRef `json:"globalAccess,omitempty"`
	// ReferenceName is an optional static name the object should be
	// use in a local repository context. It is use by a repository
	// to optionally determine a globally referencable access according
	// to the OCI distribution spec. The result will be stored
	// by the repository in the field ImageReference.
	// The value is typically an OCI repository name optionally
	// followed by a colon ':' and a tag
	ReferenceName string `json:"referenceName,omitempty"`
}

type converterV1 struct{}

var LocalBlobV1 = cpi.NewAccessSpecVersion(&AccessSpecV1{}, converterV1{})

func (_ converterV1) ConvertFrom(object cpi.AccessSpec) (runtime.TypedObject, error) {
	in := object.(*AccessSpec)
	return &AccessSpecV1{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		LocalReference:      in.LocalReference,
		ReferenceName:       in.ReferenceName,
		GlobalAccess:        core.NewAccessSpecRef(in.GlobalAccess),
		MediaType:           in.MediaType,
	}, nil
}

func (_ converterV1) ConvertTo(object interface{}) (cpi.AccessSpec, error) {
	in := object.(*AccessSpecV1)
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		LocalReference:      in.LocalReference,
		ReferenceName:       in.ReferenceName,
		GlobalAccess:        in.GlobalAccess,
		MediaType:           in.MediaType,
	}, nil
}
