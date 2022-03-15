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
	"io"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/runtime"
)

// LocalFilesystemBlobType is the access type of a blob in a local filesystem.
const LocalFilesystemBlobType = "localFilesystemBlob"
const LocalFilesystemBlobTypeV1 = LocalFilesystemBlobType + "/v1"

// Keep old access method and map generic one to this implementation for component archives

func init() {
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(LocalFilesystemBlobType, LocalFilesystemBlobV1))
	cpi.RegisterAccessType(cpi.NewConvertedAccessSpecType(LocalFilesystemBlobTypeV1, LocalFilesystemBlobV1))
}

// NewLocalFilesystemBlobAccessSpecV1 creates a new localFilesystemBlob accessor.
func NewLocalFilesystemBlobAccessSpecV1(path string, mediaType string) *accessmethods.LocalBlobAccessSpec {
	return &accessmethods.LocalBlobAccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(LocalFilesystemBlobType),
		ReferenceName:       path,
		MediaType:           mediaType,
	}
}

// LocalFilesystemBlobAccessSpec describes the access for a blob on the filesystem.
type LocalFilesystemBlobAccessSpecV1 struct {
	runtime.ObjectVersionedType `json:",inline"`
	// FileName is the
	Filename string `json:"fileName"`
	// MediaType is the media type of the object represented by the blob
	MediaType string `json:"mediaType"`
}

////////////////////////////////////////////////////////////////////////////////

type localfsblobConverterV1 struct{}

var LocalFilesystemBlobV1 = cpi.NewAccessSpecVersion(&LocalFilesystemBlobAccessSpecV1{}, localfsblobConverterV1{})

func (_ localfsblobConverterV1) ConvertFrom(object cpi.AccessSpec) (runtime.TypedObject, error) {
	in := object.(*accessmethods.LocalBlobAccessSpec)
	return &LocalFilesystemBlobAccessSpecV1{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		Filename:            in.ReferenceName,
		MediaType:           in.MediaType,
	}, nil
}

func (_ localfsblobConverterV1) ConvertTo(object interface{}) (cpi.AccessSpec, error) {
	in := object.(*LocalFilesystemBlobAccessSpecV1)
	return &accessmethods.LocalBlobAccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(in.Type),
		ReferenceName:       in.Filename,
		MediaType:           in.MediaType,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type localFilesystemBlobAccessMethod struct {
	spec *accessmethods.LocalBlobAccessSpec
	comp *ComponentVersionAccess
}

var _ cpi.AccessMethod = (*localFilesystemBlobAccessMethod)(nil)

func newLocalFilesystemBlobAccessMethod(a *accessmethods.LocalBlobAccessSpec, comp *ComponentVersionAccess) (cpi.AccessMethod, error) {
	return &localFilesystemBlobAccessMethod{
		spec: a,
		comp: comp,
	}, nil
}

func (m *localFilesystemBlobAccessMethod) GetKind() string {
	return accessmethods.LocalBlobType
}

func (m *localFilesystemBlobAccessMethod) Reader() (io.ReadCloser, error) {
	return accessio.BlobReader(m.comp.base.GetBlobData(m.spec.LocalReference))
}

func (m *localFilesystemBlobAccessMethod) Get() ([]byte, error) {
	return accessio.BlobData(m.comp.base.GetBlobData(m.spec.LocalReference))
}

func (m *localFilesystemBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
