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

package ocireg

import (
	"fmt"
	"io"
	"os"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const ArtefactsDir = "/"

var RepoFormatDirectory = DirectoryRepositoryHandler{}

func init() {
	RegisterRepositoryFormat(RepoFormatDirectory)
}

type DirectoryRepositoryHandler struct{}

// ApplyOption applies the configured path filesystem.
func (o DirectoryRepositoryHandler) ApplyOption(options *CTFOptions) {
	f := o.Format()
	options.FileFormat = &f
}

func (_ DirectoryRepositoryHandler) Format() accessio.FileFormat {
	return accessio.FormatDirectory
}

func (_ DirectoryRepositoryHandler) Open(ctx cpi.Context, path string, opts CTFOptions) (*Repository, error) {
	fs, err := projectionfs.New(opts.PathFileSystem, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projected filesystem from path %s: %w", path, err)
	}
	opts.Representation = fs // TODO: use of temporary copy
	return NewRepository(ctx, fs, nil, os.ModePerm)
}

func (_ DirectoryRepositoryHandler) Create(ctx cpi.Context, path string, opts CTFOptions, mode os.FileMode) (*Repository, error) {
	err := opts.PathFileSystem.Mkdir(path, 0660)
	if err != nil {
		return nil, err
	}
	opts.Representation, err = projectionfs.New(opts.PathFileSystem, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projected filesystem from path %s: %w", path, err)
	}
	return NewRepository(ctx, opts.Representation, nil, mode)
}

// WriteToFilesystem writes the current component archive to a filesystem
func (_ DirectoryRepositoryHandler) Write(repo *Repository, path string, opts CTFOptions, mode os.FileMode) error {
	err := opts.PathFileSystem.MkdirAll(filepath.Join(path, ArtefactsDir), mode|0400)
	if err != nil {
		return err
	}

	// copy all artefacts
	archInfos, err := vfs.ReadDir(repo.fs, ArtefactsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read artefact archives: %w", err)
	}
	for _, archInfo := range archInfos {
		if archInfo.IsDir() {
			continue
		}
		inpath := ArchPath(archInfo.Name())
		outpath := filepath.Join(path, inpath)
		arch, err := repo.fs.Open(inpath)
		if err != nil {
			return fmt.Errorf("unable to open input archive %q: %w", inpath, err)
		}
		out, err := opts.PathFileSystem.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return fmt.Errorf("unable to open output archive %q: %w", outpath, err)
		}
		if _, err := io.Copy(out, arch); err != nil {
			return fmt.Errorf("unable to copy archive from %q to %q: %w", inpath, outpath, err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("unable to close output archive %s: %w", outpath, err)
		}
		if err := arch.Close(); err != nil {
			return fmt.Errorf("unable to close input archive %s: %w", outpath, err)
		}
	}

	return nil
}

func ArchPath(name string) string {
	return filepath.Join(ArtefactsDir, name)
}
