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

package accessobj

import (
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

type FilesystemSetup func(fs vfs.FileSystem, mode vfs.FileMode) error

// InternalRepresentationFilesystem defaults a filesystem to temp filesystem and adapts.
func InternalRepresentationFilesystem(acc AccessMode, fs vfs.FileSystem, setup FilesystemSetup, mode vfs.FileMode) (bool, vfs.FileSystem, error) {
	var err error

	tmp := false
	if fs == nil {
		fs, err = osfs.NewTempFileSystem()
		if err != nil {
			return false, nil, err
		}
		tmp = true
	}
	if !acc.IsReadonly() && setup != nil {
		err = setup(fs, mode)
		if err != nil {
			return false, nil, err
		}
	}
	return tmp, fs, err
}

func HandleAccessMode(acc AccessMode, path string, opts accessio.Options, olist ...accessio.Option) (accessio.Options, bool, error) {
	ok := true
	o, err := accessio.AccessOptions(opts, olist...)
	if err != nil {
		return nil, false, err
	}
	if o.GetFile() == nil && o.GetReader() == nil {
		ok, err = vfs.Exists(o.GetPathFileSystem(), path)
		if err != nil {
			return o, false, err
		}
	}
	if !ok {
		if !acc.IsCreate() {
			return o, false, errors.ErrNotFoundWrap(vfs.ErrNotExist, "file", path)
		}
		if o.GetFileFormat() == nil {
			o.SetFileFormat(accessio.FormatDirectory)
		}
		return o, true, nil
	}

	err = o.DefaultForPath(path)
	return o, false, err
}
