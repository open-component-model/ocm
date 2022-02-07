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
	"fmt"
	"io"
	"os"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/format"
	"github.com/gardener/ocm/pkg/utils"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var FormatTAR = TarHandler{}

func init() {
	RegisterFormat(FormatTAR)
}

type TarHandler struct{}

// ApplyOption applies the configured path filesystem.
func (o TarHandler) ApplyOption(options *Options) {
	f := o.Format()
	options.FileFormat = &f
}

func (_ TarHandler) Format() accessio.FileFormat {
	return accessio.FormatTar
}

func (c TarHandler) Open(info *AccessObjectInfo, acc AccessMode, path string, opts Options) (*AccessObject, error) {
	if err := opts.ValidForPath(path); err != nil {
		return nil, err
	}
	var file vfs.File
	var err error
	if opts.File == nil {
		// we expect that the path point to a tar
		file, err = opts.PathFileSystem.Open(path)
		if err != nil {
			return nil, fmt.Errorf("unable to open tar archive from %s: %w", path, err)
		}
		defer file.Close()
	} else {
		file = opts.File
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return newFromTarReader(info, acc, file, opts, CloserFunction(func(obj *AccessObject) error { return c.close(obj, path, opts, fi.Mode()) }))
}

func (c TarHandler) Create(info *AccessObjectInfo, path string, opts Options, mode vfs.FileMode) (*AccessObject, error) {
	if err := opts.ValidForPath(path); err != nil {
		return nil, err
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

	return NewAccessObject(info, ACC_CREATE, opts.Representation, CloserFunction(func(obj *AccessObject) error { return c.close(obj, path, opts, mode) }), format.DirMode)
}

// Write tars the current descriptor and its artifacts.
func (c TarHandler) Write(obj *AccessObject, path string, opts Options, mode vfs.FileMode) error {
	writer, err := opts.WriterFor(path, mode)
	if err != nil {
		return err
	}
	return c.WriteToStream(obj, writer, opts)
}

func (_ TarHandler) WriteToStream(obj *AccessObject, writer io.Writer, opts Options) error {
	// write descriptor
	err := obj.Update()
	if err != nil {
		return err
	}
	data, err := obj.state.GetBlob()
	if err != nil {
		return err
	}
	tw := tar.NewWriter(writer)
	cdHeader := &tar.Header{
		Name:    obj.info.DescriptorFileName,
		Size:    data.Size(),
		Mode:    format.FileMode,
		ModTime: format.ModTime,
	}

	if err := tw.WriteHeader(cdHeader); err != nil {
		return fmt.Errorf("unable to write descriptor header: %w", err)
	}
	r, err := data.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	if _, err := io.Copy(tw, r); err != nil {
		return fmt.Errorf("unable to write descriptor content: %w", err)
	}

	// add all content
	err = tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     obj.info.ElementDirectoryName,
		Mode:     format.DirMode,
		ModTime:  format.ModTime,
	})
	if err != nil {
		return fmt.Errorf("unable to write %s directory: %w", obj.info.ElementTypeName, err)
	}

	fileInfos, err := vfs.ReadDir(obj.fs, obj.info.ElementDirectoryName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read %s directory: %w", obj.info.ElementTypeName, err)
	}
	for _, fileInfo := range fileInfos {
		path := obj.info.SubPath(fileInfo.Name())
		header := &tar.Header{
			Name:    path,
			Size:    fileInfo.Size(),
			Mode:    format.FileMode,
			ModTime: format.ModTime,
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write %s header: %w", obj.info.ElementTypeName, err)
		}

		content, err := obj.fs.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open %s: %w", obj.info.ElementTypeName, err)
		}
		if _, err := io.Copy(tw, content); err != nil {
			return fmt.Errorf("unable to write %s content: %w", obj.info.ElementTypeName, err)
		}
		if err := content.Close(); err != nil {
			return fmt.Errorf("unable to close %s %s: %w", obj.info.ElementTypeName, path, err)
		}
	}

	return tw.Close()
}

func (c TarHandler) close(obj *AccessObject, path string, opts Options, mode vfs.FileMode) error {
	return c.Write(obj, path, opts, mode)
}

// newFromTarReader creates a new builder from a input reader.
func newFromTarReader(info *AccessObjectInfo, acc AccessMode, in io.Reader, opts Options, closer Closer) (*AccessObject, error) {
	if err := utils.ExtractTarToFs(opts.Representation, in); err != nil {
		return nil, fmt.Errorf("unable to extract tar: %w", err)
	}

	return NewAccessObject(info, acc, opts.Representation, closer, format.DirMode)
}
