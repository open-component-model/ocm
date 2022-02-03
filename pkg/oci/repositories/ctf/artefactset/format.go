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

package artefactset

import (
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/format"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const ArtefactSetDescriptorFileName = format.ArtefactSetDescriptorFileName
const BlobsDirectoryName = format.BlobsDirectoryName

var accessObjectInfo = &accessobj.AccessObjectInfo{
	DescriptorFileName:       ArtefactSetDescriptorFileName,
	ObjectTypeName:           "artefactset",
	ElementDirectoryName:     BlobsDirectoryName,
	ElementTypeName:          "blob",
	DescriptorHandlerFactory: NewStateHandler,
}

type Object = ArtefactSet

type FormatHandler struct {
	accessobj.FormatHandler
}

var (
	FormatDirectory = RegisterFormat(accessobj.FormatDirectory)
	FormatTAR       = RegisterFormat(accessobj.FormatTAR)
	FormatTGZ       = RegisterFormat(accessobj.FormatTGZ)
)

////////////////////////////////////////////////////////////////////////////////

var fileFormats = map[accessio.FileFormat]FormatHandler{}
var lock sync.RWMutex

func RegisterFormat(f accessobj.FormatHandler) FormatHandler {
	lock.Lock()
	defer lock.Unlock()
	h := FormatHandler{f}
	fileFormats[f.Format()] = h
	return h
}

func GetFormat(name accessio.FileFormat) FormatHandler {
	lock.RLock()
	defer lock.RUnlock()
	return fileFormats[name]
}

////////////////////////////////////////////////////////////////////////////////

func Open(acc accessobj.AccessMode, path string, opts ...accessobj.Option) (*Object, error) {
	o := accessobj.AccessOptions(opts...)
	h, ok := fileFormats[*o.FileFormat]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.FileFormat.String())
	}

	ok, err := vfs.Exists(o.PathFileSystem, path)
	if err != nil {
		return nil, err
	}
	if !ok && acc.IsCreate() {
		return h.Create(path, o, 0700)
	}
	return h.Open(acc, path, o)
}

func Create(acc accessobj.AccessMode, path string, mode vfs.FileMode, opts ...accessobj.Option) (*Object, error) {
	o := accessobj.AccessOptions(opts...)
	h, ok := fileFormats[*o.FileFormat]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.FileFormat.String())
	}
	return h.Create(path, o, mode)
}

////////////////////////////////////////////////////////////////////////////////

func (h FormatHandler) Open(acc accessobj.AccessMode, path string, opts accessobj.Options) (*Object, error) {

	return _Wrap(h.FormatHandler.Open(accessObjectInfo, acc, path, opts))
}

func (h *FormatHandler) Create(path string, opts accessobj.Options, mode vfs.FileMode) (*Object, error) {
	return _Wrap(h.FormatHandler.Create(accessObjectInfo, path, opts, mode))
}

// WriteToFilesystem writes the current object to a filesystem
func (h *FormatHandler) Write(obj *Object, path string, opts accessobj.Options, mode vfs.FileMode) error {
	return h.FormatHandler.Write(obj.base.Access(), path, opts, mode)
}
