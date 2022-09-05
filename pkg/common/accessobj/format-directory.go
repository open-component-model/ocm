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
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

var FormatDirectory = DirectoryHandler{}

func init() {
	RegisterFormat(FormatDirectory)
}

type DirectoryHandler struct{}

// ApplyOption applies the configured path filesystem.
func (o DirectoryHandler) ApplyOption(options *accessio.Options) {
	f := o.Format()
	options.FileFormat = &f
}

func (_ DirectoryHandler) Format() accessio.FileFormat {
	return accessio.FormatDirectory
}

func (_ DirectoryHandler) Open(info *AccessObjectInfo, acc AccessMode, path string, opts accessio.Options) (*AccessObject, error) {
	if err := opts.ValidForPath(path); err != nil {
		return nil, err
	}
	if opts.File != nil || opts.Reader != nil {
		return nil, errors.ErrNotSupported("file or reader option")
	}
	fs, err := projectionfs.New(opts.PathFileSystem, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projected filesystem from path %s: %w", path, err)
	}
	opts.Representation = fs // TODO: use of temporary copy
	return NewAccessObject(info, acc, fs, nil, nil, os.ModePerm)
}

func (_ DirectoryHandler) Create(info *AccessObjectInfo, path string, opts accessio.Options, mode vfs.FileMode) (*AccessObject, error) {
	if err := opts.ValidForPath(path); err != nil {
		return nil, err
	}
	if opts.File != nil || opts.Reader != nil {
		return nil, errors.ErrNotSupported("file or reader option")
	}
	err := opts.PathFileSystem.Mkdir(path, mode)
	if err != nil {
		return nil, err
	}
	opts.Representation, err = projectionfs.New(opts.PathFileSystem, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projected filesystem from path %s: %w", path, err)
	}
	return NewAccessObject(info, ACC_CREATE, opts.Representation, nil, nil, mode)
}

// WriteToFilesystem writes the current object to a filesystem.
func (_ DirectoryHandler) Write(obj *AccessObject, path string, opts accessio.Options, mode vfs.FileMode) error {
	// create the directory structure with the content directory
	if err := opts.PathFileSystem.MkdirAll(filepath.Join(path, obj.info.ElementDirectoryName), mode|0o400); err != nil {
		return fmt.Errorf("unable to create output directory %q: %w", path, err)
	}

	_, err := obj.updateDescriptor()
	if err != nil {
		return fmt.Errorf("unable to update descriptor: %w", err)
	}

	// copy descriptor
	err = vfs.CopyFile(obj.fs, obj.info.DescriptorFileName, opts.PathFileSystem, filepath.Join(path, obj.info.DescriptorFileName))
	if err != nil {
		return fmt.Errorf("unable to copy file '%s': %w", obj.info.DescriptorFileName, err)
	}

	// copy all content
	fileInfos, err := vfs.ReadDir(obj.fs, obj.info.ElementDirectoryName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read '%s': %w", obj.info.ElementDirectoryName, err)
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		inpath := obj.info.SubPath(fileInfo.Name())
		outpath := filepath.Join(path, inpath)
		content, err := obj.fs.Open(inpath)
		if err != nil {
			return fmt.Errorf("unable to open input %s %q: %w", obj.info.ElementTypeName, inpath, err)
		}
		out, err := opts.PathFileSystem.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode|0o666)
		if err != nil {
			return fmt.Errorf("unable to open output %s %q: %w", obj.info.ElementTypeName, outpath, err)
		}
		if _, err := io.Copy(out, content); err != nil {
			return fmt.Errorf("unable to copy %s from %q to %q: %w", obj.info.ElementTypeName, inpath, outpath, err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("unable to close output %s %s: %w", obj.info.ElementTypeName, outpath, err)
		}
		if err := content.Close(); err != nil {
			return fmt.Errorf("unable to close input %s %s: %w", obj.info.ElementTypeName, outpath, err)
		}
	}

	return nil
}
