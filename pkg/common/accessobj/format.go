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
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/format"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const KIND_FILEFORMAT = "file format"

type FileFormat = accessio.FileFormat

type FormatHandler interface {
	Option

	Format() accessio.FileFormat

	Open(info *AccessObjectInfo, acc AccessMode, path string, opts Options) (*AccessObject, error)
	Create(info *AccessObjectInfo, path string, opts Options, mode vfs.FileMode) (*AccessObject, error)
	Write(obj *AccessObject, path string, opts Options, mode vfs.FileMode) error
}

////////////////////////////////////////////////////////////////////////////////

var fileFormats = map[FileFormat]FormatHandler{}
var lock sync.RWMutex

func RegisterFormat(f FormatHandler) {
	lock.Lock()
	defer lock.Unlock()
	fileFormats[f.Format()] = f
}

func GetFormat(name FileFormat) FormatHandler {
	lock.RLock()
	defer lock.RUnlock()
	return fileFormats[name]
}

////////////////////////////////////////////////////////////////////////////////

type Closer interface {
	Close(*AccessObject) error
}

type CloserFunction func(*AccessObject) error

func (f CloserFunction) Close(obj *AccessObject) error {
	return f(obj)
}

////////////////////////////////////////////////////////////////////////////////

type Setup interface {
	Setup(vfs.FileSystem) error
}

type SetupFunction func(vfs.FileSystem) error

func (f SetupFunction) Setup(fs vfs.FileSystem) error {
	return f(fs)
}

////////////////////////////////////////////////////////////////////////////////

type fsCloser struct {
	closer Closer
}

func FSCloser(closer Closer) Closer {
	return &fsCloser{closer}
}

func (f fsCloser) Close(obj *AccessObject) error {
	err := errors.ErrListf("cannot close %s", obj.info.ObjectTypeName)
	if f.closer != nil {
		err.Add(f.closer.Close(obj))
	}
	err.Add(vfs.Cleanup(obj.fs))
	return err.Result()
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

	format := accessio.FormatDirectory
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

	format := accessio.FormatDirectory
	if !fi.IsDir() {
		var r io.Reader

		defer file.Seek(0, io.SeekStart)
		zip, err := gzip.NewReader(file)
		if err == nil {
			format = accessio.FormatTGZ
			defer zip.Close()
			r = zip
		} else {
			file.Seek(0, io.SeekStart)
			format = accessio.FormatTar
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

type StandardReaderHandler interface {
	Write(obj *AccessObject, path string, opts Options, mode vfs.FileMode) error
	NewFromReader(info *AccessObjectInfo, acc AccessMode, in io.Reader, opts Options, closer Closer) (*AccessObject, error)
}

func DefaultOpenOptsFileHandling(kind string, info *AccessObjectInfo, acc AccessMode, path string, opts Options, handler StandardReaderHandler) (*AccessObject, error) {
	if err := opts.ValidForPath(path); err != nil {
		return nil, err
	}
	var reader io.ReadCloser
	var file vfs.File
	var err error
	var closer Closer
	if opts.Reader != nil {
		reader = opts.Reader
		defer opts.Reader.Close()
	} else if opts.File == nil {
		// we expect that the path point to a tar
		file, err = opts.PathFileSystem.Open(path)
		if err != nil {
			return nil, fmt.Errorf("unable to open %s from %s: %w", kind, path, err)
		}
		defer file.Close()
	} else {
		file = opts.File
	}
	if file != nil {
		reader = file
		fi, err := file.Stat()
		if err != nil {
			return nil, err
		}
		closer = CloserFunction(func(obj *AccessObject) error { return handler.Write(obj, path, opts, fi.Mode()) })
	}
	return handler.NewFromReader(info, acc, reader, opts, closer)
}

func DefaultCreateOptsFileHandling(kind string, info *AccessObjectInfo, path string, opts Options, mode vfs.FileMode, handler StandardReaderHandler) (*AccessObject, error) {
	if err := opts.ValidForPath(path); err != nil {
		return nil, err
	}
	if opts.Reader != nil {
		return nil, errors.ErrNotSupported("reader option not supported")
	}
	if opts.File == nil {
		ok, err := vfs.Exists(opts.PathFileSystem, path)
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, vfs.ErrExist
		}
	}

	return NewAccessObject(info, ACC_CREATE, opts.Representation, nil, CloserFunction(func(obj *AccessObject) error { return handler.Write(obj, path, opts, mode) }), format.DirMode)
}
