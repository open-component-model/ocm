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

package localfsblob

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type of a blob in a local filesystem.
const Type = "localFilesystemBlob"
const TypeV1 = Type + "/v1"

// Keep old access method and map generic one to this implementation for component archives

func init() {
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(Type, LocalFilesystemBlobV1))
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(TypeV1, LocalFilesystemBlobV1))
}

// New creates a new localFilesystemBlob accessor.
func New(path string, media string) *localblob.AccessSpec {
	return &localblob.AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		LocalReference:      path,
		MediaType:           media,
	}
}

// AccessSpec describes the access for a blob on the filesystem.
// Deprecated: use LocalBlob.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	// FileName is the
	Filename string `json:"fileName"`
	// MediaType is the media type of the object represented by the blob
	MediaType string `json:"mediaType"`
}

////////////////////////////////////////////////////////////////////////////////

type localfsblobConverterV1 struct{}

var LocalFilesystemBlobV1 = cpi.NewAccessSpecVersion(&AccessSpec{}, localfsblobConverterV1{})

func (_ localfsblobConverterV1) ConvertFrom(object cpi.AccessSpec) (runtime.TypedObject, error) {
	in := object.(*localblob.AccessSpec)
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		Filename:            in.LocalReference,
		MediaType:           in.MediaType,
	}, nil
}

func (_ localfsblobConverterV1) ConvertTo(object interface{}) (cpi.AccessSpec, error) {
	in := object.(*AccessSpec)
	return &localblob.AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		LocalReference:      in.Filename,
		MediaType:           in.MediaType,
	}, nil
}
