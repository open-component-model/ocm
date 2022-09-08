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

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/format"
	"github.com/open-component-model/ocm/pkg/errors"
	mime2 "github.com/open-component-model/ocm/pkg/mime"
)

const (
	ArtefactSetDescriptorFileName = format.ArtefactSetDescriptorFileName
	BlobsDirectoryName            = format.BlobsDirectoryName

	OCIArtefactSetDescriptorFileName = "index.json"
	OCILayouFileName                 = "oci-layout"
)

var DefaultArtefactSetDescriptorFileName = OCIArtefactSetDescriptorFileName

func IsOCIDefaultFormat() bool {
	return DefaultArtefactSetDescriptorFileName == OCIArtefactSetDescriptorFileName
}

type accessObjectInfo struct {
	accessobj.DefaultAccessObjectInfo
}

var _ accessobj.AccessObjectInfo = (*accessObjectInfo)(nil)

func NewAccessObjectInfo(fmts ...string) accessobj.AccessObjectInfo {
	a := &accessObjectInfo{
		accessobj.DefaultAccessObjectInfo{
			ObjectTypeName:           "artefactset",
			ElementDirectoryName:     BlobsDirectoryName,
			ElementTypeName:          "blob",
			DescriptorHandlerFactory: NewStateHandler,
		},
	}
	oci := IsOCIDefaultFormat()
	if len(fmts) > 0 {
		switch fmts[0] {
		case FORMAT_OCM:
			oci = false
		case FORMAT_OCI:
			oci = true
		case "":
		}
	}
	if oci {
		a.setOCI()
	} else {
		a.setOCM()
	}
	return a
}

func (a *accessObjectInfo) setOCI() {
	a.DescriptorFileName = OCIArtefactSetDescriptorFileName
	a.AdditionalFiles = []string{OCILayouFileName}
}

func (a *accessObjectInfo) setOCM() {
	a.DescriptorFileName = ArtefactSetDescriptorFileName
	a.AdditionalFiles = nil
}

func (a *accessObjectInfo) setupOCIFS(fs vfs.FileSystem, mode vfs.FileMode) error {
	data := `{
    "imageLayoutVersion": "1.0.0"
}
`
	return vfs.WriteFile(fs, OCILayouFileName, []byte(data), mode)
}

func (a *accessObjectInfo) SetupFileSystem(fs vfs.FileSystem, mode vfs.FileMode) error {
	if err := a.SetupFor(fs); err != nil {
		return err
	}
	if err := a.DefaultAccessObjectInfo.SetupFileSystem(fs, mode); err != nil {
		return err
	}
	if len(a.AdditionalFiles) > 0 {
		return a.setupOCIFS(fs, mode)
	}
	return nil
}

func (a *accessObjectInfo) SetupFor(fs vfs.FileSystem) error {
	ok, err := vfs.FileExists(fs, OCILayouFileName)
	if err != nil {
		return err
	}
	if ok {
		a.setOCI()
	} else { //nolint: staticcheck // keep comment for else
		// keep configured format
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Object = ArtefactSet

type FormatHandler interface {
	accessio.Option

	Format() accessio.FileFormat

	Open(acc accessobj.AccessMode, path string, opts accessio.Options) (*Object, error)
	Create(path string, opts accessio.Options, mode vfs.FileMode) (*Object, error)
	Write(obj *Object, path string, opts accessio.Options, mode vfs.FileMode) error
}

type formatHandler struct {
	accessobj.FormatHandler
}

var (
	FormatDirectory = RegisterFormat(accessobj.FormatDirectory)
	FormatTAR       = RegisterFormat(accessobj.FormatTAR)
	FormatTGZ       = RegisterFormat(accessobj.FormatTGZ)
)

////////////////////////////////////////////////////////////////////////////////

var (
	fileFormats = map[accessio.FileFormat]FormatHandler{}
	lock        sync.RWMutex
)

func RegisterFormat(f accessobj.FormatHandler) FormatHandler {
	lock.Lock()
	defer lock.Unlock()
	h := &formatHandler{f}
	fileFormats[f.Format()] = h
	return h
}

func GetFormat(name accessio.FileFormat) FormatHandler {
	lock.RLock()
	defer lock.RUnlock()
	return fileFormats[name]
}

func SupportedFormats() []accessio.FileFormat {
	lock.RLock()
	defer lock.RUnlock()
	result := make([]accessio.FileFormat, 0, len(fileFormats))
	for f := range fileFormats {
		result = append(result, f)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////

func OpenFromBlob(acc accessobj.AccessMode, blob accessio.BlobAccess, opts ...accessio.Option) (*Object, error) {
	o, err := accessio.AccessOptions(nil, opts...)
	if err != nil {
		return nil, err
	}
	if o.GetFile() != nil || o.GetReader() != nil {
		return nil, errors.ErrInvalid("file or reader option nor possible for blob access")
	}
	reader, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	o.SetReader(reader)
	fmt := accessio.FormatTar
	mime := blob.MimeType()

	if mime2.IsGZip(mime) {
		fmt = accessio.FormatTGZ
	}
	o.SetFileFormat(fmt)
	return Open(acc&accessobj.ACC_READONLY, "", 0, o)
}

func Open(acc accessobj.AccessMode, path string, mode vfs.FileMode, olist ...accessio.Option) (*Object, error) {
	o, create, err := accessobj.HandleAccessMode(acc, path, &Options{}, olist...)
	if err != nil {
		return nil, err
	}
	h, ok := fileFormats[*o.GetFileFormat()]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.GetFileFormat().String())
	}
	if create {
		return h.Create(path, o, mode)
	}
	return h.Open(acc, path, o)
}

func Create(acc accessobj.AccessMode, path string, mode vfs.FileMode, opts ...accessio.Option) (*Object, error) {
	o, err := accessio.AccessOptions(&Options{}, opts...)
	if err != nil {
		return nil, err
	}
	o.DefaultFormat(accessio.FormatDirectory)
	h, ok := fileFormats[*o.GetFileFormat()]
	if !ok {
		return nil, errors.ErrUnknown(accessobj.KIND_FILEFORMAT, o.GetFileFormat().String())
	}
	return h.Create(path, o, mode)
}

////////////////////////////////////////////////////////////////////////////////

func (h *formatHandler) Open(acc accessobj.AccessMode, path string, opts accessio.Options) (*Object, error) {
	return _Wrap(h.FormatHandler.Open(NewAccessObjectInfo(GetFormatVersion(opts)), acc, path, opts))
}

func (h *formatHandler) Create(path string, opts accessio.Options, mode vfs.FileMode) (*Object, error) {
	return _Wrap(h.FormatHandler.Create(NewAccessObjectInfo(GetFormatVersion(opts)), path, opts, mode))
}

// WriteToFilesystem writes the current object to a filesystem.
func (h *formatHandler) Write(obj *Object, path string, opts accessio.Options, mode vfs.FileMode) error {
	return h.FormatHandler.Write(obj.base.Access(), path, opts, mode)
}
