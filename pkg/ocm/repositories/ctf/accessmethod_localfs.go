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
	"fmt"
	"io"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/core/accesstypes"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

// LocalFilesystemBlobType is the access type of a blob in a local filesystem.
const LocalFilesystemBlobType = "localFilesystemBlob"
const LocalFilesystemBlobTypeV1 = LocalFilesystemBlobType + "/v1"

// Keep old access method and map generic one to this implementation for component archives

func init() {
	core.RegisterAccessType(accesstypes.NewConvertedType(LocalFilesystemBlobType, LocalFilesystemBlobV1))
	core.RegisterAccessType(accesstypes.NewConvertedType(LocalFilesystemBlobTypeV1, LocalFilesystemBlobV1))
}

// NewLocalFilesystemBlobAccessSpecV1 creates a new localFilesystemBlob accessor.
func NewLocalFilesystemBlobAccessSpecV1(path string, mediaType string) *LocalFilesystemBlobAccessSpec {
	return &LocalFilesystemBlobAccessSpec{
		LocalBlobAccessSpec: accessmethods.LocalBlobAccessSpec{
			ObjectTypeVersion: runtime.NewObjectTypeVersion(LocalFilesystemBlobType),
			Filename:          path,
			MediaType:         mediaType,
		},
	}
}

// LocalFilesystemBlobAccessSpec describes the access for a blob on the filesystem.
type LocalFilesystemBlobAccessSpec struct {
	accessmethods.LocalBlobAccessSpec `json:",inline"`
}

func (a *LocalFilesystemBlobAccessSpec) ValidFor(repo core.Repository) bool {
	return repo.GetSpecification().GetName() == CTFRepositoryType
}

func (a *LocalFilesystemBlobAccessSpec) AccessMethod(c core.ComponentAccess) (core.AccessMethod, error) {
	rtype := c.GetAccessType()
	if rtype != CTFRepositoryType {
		return nil, errors.ErrNotSupported(errors.KIND_ACCESSMETHOD, c.GetName(), rtype)
	}
	acc, ok := c.(*ComponentArchive)
	if !ok {
		return nil, fmt.Errorf("implementation error: expected type ComponentArchive but got %T", c)
	}
	return newLocalFilesystemBlobAccessMethod(a, acc)
}

////////////////////////////////////////////////////////////////////////////////

type localfsblobConverterV1 struct{}

var LocalFilesystemBlobV1 = accesstypes.NewAccessSpecVersion(&accessmethods.LocalBlobAccessSpecV1{}, localfsblobConverterV1{})

func (_ localfsblobConverterV1) ConvertFrom(object core.AccessSpec) (runtime.TypedObject, error) {
	in := object.(*LocalFilesystemBlobAccessSpec)
	return &accessmethods.LocalBlobAccessSpecV1{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(in.Type),
		Filename:          in.Filename,
		MediaType:         in.MediaType,
	}, nil
}

func (_ localfsblobConverterV1) ConvertTo(object interface{}) (core.AccessSpec, error) {
	in := object.(*accessmethods.LocalBlobAccessSpecV1)
	return &LocalFilesystemBlobAccessSpec{
		LocalBlobAccessSpec: accessmethods.LocalBlobAccessSpec{
			ObjectTypeVersion: runtime.NewObjectTypeVersion(in.Type),
			Filename:          in.Filename,
			MediaType:         in.MediaType,
		},
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type localFilesystemBlobAccessMethod struct {
	spec *LocalFilesystemBlobAccessSpec
	comp *ComponentArchive
}

var _ accessmethods.AccessImplementation = &localFilesystemBlobAccessMethod{}

func newLocalFilesystemBlobAccessMethod(a *LocalFilesystemBlobAccessSpec, comp *ComponentArchive) (core.AccessMethod, error) {
	return accessmethods.NewDefaultAccessMethod(LocalFilesystemBlobType, &localFilesystemBlobAccessMethod{
		spec: a,
		comp: comp,
	}), nil
}

func (m *localFilesystemBlobAccessMethod) Open() (io.ReadCloser, error) {
	blobpath := BlobPath(m.spec.Filename)

	info, err := m.comp.fs.Stat(blobpath)
	if err != nil {
		return nil, fmt.Errorf("unable to get fileinfo for %s: %w", blobpath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("directories are not allowed as blobs %s", blobpath)
	}
	file, err := m.comp.fs.Open(blobpath)
	if err != nil {
		return nil, fmt.Errorf("unable to open blob from %s", blobpath)
	}
	return file, nil
}

func (m *localFilesystemBlobAccessMethod) Size() int64 {
	info, err := m.comp.fs.Stat(BlobPath(m.spec.Filename))
	if err == nil {
		return info.Size()
	}
	return -1
}

func (m *localFilesystemBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
