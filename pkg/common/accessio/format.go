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

package accessio

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/errors"
)

const KIND_FILEFORMAT = "file format"

type FileFormat string

func (f FileFormat) String() string {
	return string(f)
}

func (o FileFormat) ApplyOption(options Options) error {
	if o != "" {
		options.SetFileFormat(o)
	}
	return nil
}

const (
	FormatTar       FileFormat = "tar"
	FormatTGZ       FileFormat = "tgz"
	FormatDirectory FileFormat = "directory"
)

func ErrInvalidFileFormat(fmt string) error {
	return errors.ErrInvalid(KIND_FILEFORMAT, fmt)
}

////////////////////////////////////////////////////////////////////////////////

func FileFormatForType(t string) FileFormat {
	i := strings.Index(t, "+")
	if i < 0 {
		return FileFormat(t)
	}
	return FileFormat(t[i+1:])
}

func TypeForType(t string) string {
	i := strings.Index(t, "+")
	if i < 0 {
		return ""
	}
	return t[:i]
}

////////////////////////////////////////////////////////////////////////////////

func DetectFormat(path string, fs vfs.FileSystem) (*FileFormat, error) {
	if fs == nil {
		fs = _osfs
	}

	fi, err := fs.Stat(path)
	if err != nil {
		return nil, err
	}

	format := FormatDirectory
	if !fi.IsDir() {
		file, err := fs.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return DetectFormatForFile(file)
	}
	return &format, nil
}

func DetectFormatForFile(file vfs.File) (*FileFormat, error) {
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	format := FormatDirectory
	if !fi.IsDir() {
		var r io.Reader

		defer file.Seek(0, io.SeekStart)
		zip, err := gzip.NewReader(file)
		if err == nil {
			format = FormatTGZ
			defer zip.Close()
			r = zip
		} else {
			file.Seek(0, io.SeekStart)
			format = FormatTar
			r = file
		}
		t := tar.NewReader(r)
		_, err = t.Next()
		if err != nil {
			return nil, err
		}
	}
	return &format, nil
}
