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

// InternalRepresentationFilesystem defaults a filesystem to temp filesystem and adapts.
func InternalRepresentationFilesystem(acc AccessMode, fs vfs.FileSystem, dir string, mode vfs.FileMode) (bool, vfs.FileSystem, error) {
	var err error

	tmp := false
	if fs == nil {
		fs, err = osfs.NewTempFileSystem()
		if err != nil {
			return false, nil, err
		}
		tmp = true
	}
	if !acc.IsReadonly() && dir != "" {
		err = fs.MkdirAll(dir, mode)
		if err != nil {
			return false, nil, err
		}
	}
	return tmp, fs, err
}

func HandleAccessMode(acc AccessMode, path string, opts ...accessio.Option) (accessio.Options, bool, error) {
	var err error
	ok := true
	o := accessio.AccessOptions(opts...)
	if o.File == nil && o.Reader == nil {
		ok, err = vfs.Exists(o.PathFileSystem, path)
		if err != nil {
			return o, false, err
		}
	}
	if !ok {
		if !acc.IsCreate() {
			return o, false, errors.ErrNotFoundWrap(vfs.ErrNotExist, "file", path)
		}
		if o.FileFormat == nil {
			fmt := accessio.FormatDirectory
			o.FileFormat = &fmt
		}
		return o, true, nil
	}

	o, err = o.DefaultForPath(path)
	return o, false, err
}
