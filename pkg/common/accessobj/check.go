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
	"archive/tar"
	"io"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/errors"
)

func mapErr(forced bool, err error) (bool, bool, error) {
	if !forced {
		return false, false, nil
	}
	return false, true, err
}

// CheckFile returns create, acceptable, error.
func CheckFile(kind string, createHint string, forcedType bool, path string, fs vfs.FileSystem, descriptorname string) (bool, bool, error) {
	info, err := fs.Stat(path)
	if err != nil {
		if createHint == kind {
			if vfs.IsErrNotExist(err) {
				return true, true, nil
			}
		}
		return mapErr(forcedType, err)
	}
	accepted := false
	if !info.IsDir() {
		file, err := fs.Open(path)
		if err != nil {
			return mapErr(forcedType, err)
		}
		defer file.Close()
		r, _, err := compression.AutoDecompress(file)
		if err != nil {
			return mapErr(forcedType, err)
		}
		tr := tar.NewReader(r)
		for {
			header, err := tr.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return mapErr(forcedType, err)
			}

			switch header.Typeflag {
			case tar.TypeReg:
				if header.Name == descriptorname {
					accepted = true
					break
				}
			}
		}
	} else {
		if ok, err := vfs.FileExists(fs, filepath.Join(path, descriptorname)); !ok || err != nil {
			if err != nil {
				return mapErr(forcedType, err)
			}
		} else {
			accepted = ok
		}
	}
	if !accepted {
		return mapErr(forcedType, errors.Newf("%s: no %s", path, kind))
	}
	return false, true, nil
}
