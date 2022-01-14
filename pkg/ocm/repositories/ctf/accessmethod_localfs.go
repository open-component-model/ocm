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
	"io/ioutil"

	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/runtime"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

// LocalFilesystemBlobType is the access type of a blob in a local filesystem.
const LocalFilesystemBlobType = "localFilesystemBlob"
const LocalFilesystemBlobTypeV1 = LocalFilesystemBlobType + "/v1"

// NewLocalFilesystemBlobAccessSpecV1 creates a new localFilesystemBlob accessor.
func NewLocalFilesystemBlobAccessSpecV1(path string, mediaType string) *LocalFilesystemBlobAccessSpec {
	return &LocalFilesystemBlobAccessSpec{
		ObjectTypeVersion: runtime.NewObjectTypeVersion(LocalFilesystemBlobType),
		Filename:          path,
		MediaType:         mediaType,
	}
}

// LocalFilesystemBlobAccessSpec describes the access for a blob on the filesystem.
type LocalFilesystemBlobAccessSpec struct {
	runtime.ObjectTypeVersion `json:",inline"`
	// Filename is the name of the blob in the local filesystem.
	// The blob is expected to be at <fs-root>/blobs/<name>
	Filename string `json:"filename"`
	// MediaType is the media type of the object this filename refers to.
	MediaType string `json:"mediaType,omitempty"`
}

func (_ *LocalFilesystemBlobAccessSpec) GetType() string {
	return LocalFilesystemBlobType
}

func (a *LocalFilesystemBlobAccessSpec) AccessMethod(c core.ComponentAccess) (core.AccessMethod, error) {
	rtype := c.GetAccessType()
	if rtype != CTFRepositoryType {
		return nil, fmt.Errorf("access method not applicable for repository type %q", rtype)
	}
	acc, ok := c.(*ComponentArchive)
	if !ok {
		return nil, fmt.Errorf("implementation error: expected type ComponentArchive but got %T", c)
	}
	return newLocalFilesystemBlobAccessMethod(a, acc)
}

////////////////////////////////////////////////////////////////////////////////

type LocalFilesystemBlobAccessMethod struct {
	spec *LocalFilesystemBlobAccessSpec
	comp *ComponentArchive
}

var _ core.AccessMethod = &LocalFilesystemBlobAccessMethod{}

func newLocalFilesystemBlobAccessMethod(a *LocalFilesystemBlobAccessSpec, comp *ComponentArchive) (*LocalFilesystemBlobAccessMethod, error) {
	return &LocalFilesystemBlobAccessMethod{
		spec: a,
		comp: comp,
	}, nil
}

func (m *LocalFilesystemBlobAccessMethod) GetName() string {
	return LocalFilesystemBlobType
}

func (m *LocalFilesystemBlobAccessMethod) Open() (vfs.File, error) {
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

func (m *LocalFilesystemBlobAccessMethod) Get() ([]byte, error) {
	file, err := m.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

func (m *LocalFilesystemBlobAccessMethod) Reader() (io.ReadCloser, error) {
	return m.Open()
}

func (m *LocalFilesystemBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
