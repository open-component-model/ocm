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

package ctf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
	cpi "github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var CA_DIRECTORY = ComponentDirectory{}

func init() {
	RegisterComponentArchiveFormat(CA_DIRECTORY)
}

type ComponentDirectory struct{}

// ApplyOption applies the configured path filesystem.
func (o ComponentDirectory) ApplyOption(options *ComponentArchiveOptions) {
	options.ArchiveFormat = o
}

func (_ ComponentDirectory) String() string {
	return "directory"
}

func (_ ComponentDirectory) Open(ctx cpi.Context, path string, opts ComponentArchiveOptions) (*ComponentArchive, error) {
	fs, err := projectionfs.New(opts.PathFileSystem, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projected filesystem from path %s: %w", path, err)
	}
	opts.ComponentFileSystem = fs
	return newComponentArchiveFromFilesystem(ctx, opts, nil)
}

func (_ ComponentDirectory) Create(ctx cpi.Context, path string, cd *compdesc.ComponentDescriptor, opts ComponentArchiveOptions) (*ComponentArchive, error) {
	err := opts.PathFileSystem.Mkdir(path, 0660)
	if err != nil {
		return nil, err
	}
	opts.ComponentFileSystem, err = projectionfs.New(opts.PathFileSystem, path)
	if err != nil {
		return nil, fmt.Errorf("unable to create projected filesystem from path %s: %w", path, err)
	}
	return newComponentArchiveForFilesystem(ctx, cd, opts, nil)
}

// WriteToFilesystem writes the current component archive to a filesystem
func (_ ComponentDirectory) Write(ca *ComponentArchive, path string, opts ComponentArchiveOptions) error {
	// create the directory structure with the blob directory
	if err := opts.PathFileSystem.MkdirAll(filepath.Join(path, BlobsDirectoryName), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create output directory %q: %s", path, err.Error())
	}
	// copy component-descriptor
	cdBytes, err := compdesc.Encode(ca.ComponentDescriptor, opts.EncodeOptions)
	if err != nil {
		return fmt.Errorf("unable to encode component descriptor: %w", err)
	}
	if err := vfs.WriteFile(opts.PathFileSystem, filepath.Join(path, ComponentDescriptorFileName), cdBytes, os.ModePerm); err != nil {
		return fmt.Errorf("unable to copy component descritptor to %q: %w", filepath.Join(path, ComponentDescriptorFileName), err)
	}

	// copy all blobs
	blobInfos, err := vfs.ReadDir(ca.fs, BlobsDirectoryName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read blobs: %w", err)
	}
	for _, blobInfo := range blobInfos {
		if blobInfo.IsDir() {
			continue
		}
		inpath := BlobPath(blobInfo.Name())
		outpath := filepath.Join(path, BlobsDirectoryName, blobInfo.Name())
		blob, err := ca.fs.Open(inpath)
		if err != nil {
			return fmt.Errorf("unable to open input blob %q: %w", inpath, err)
		}
		out, err := opts.PathFileSystem.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to open output blob %q: %w", outpath, err)
		}
		if _, err := io.Copy(out, blob); err != nil {
			return fmt.Errorf("unable to copy blob from %q to %q: %w", inpath, outpath, err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("unable to close output blob %s: %w", outpath, err)
		}
		if err := blob.Close(); err != nil {
			return fmt.Errorf("unable to close input blob %s: %w", outpath, err)
		}
	}

	return nil
}

// newComponentArchiveFromFilesystem creates a component archive object from a filesyste,
func newComponentArchiveFromFilesystem(ctx cpi.Context, opts ComponentArchiveOptions, closer ComponentCloser) (*ComponentArchive, error) {
	data, err := vfs.ReadFile(opts.ComponentFileSystem, filepath.Join("/", ComponentDescriptorFileName))
	if err != nil {
		return nil, fmt.Errorf("unable to read the component descriptor from %s: %w", ComponentDescriptorFileName, err)
	}
	cd, err := compdesc.Decode(data, opts.DecodeOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to parse component descriptor read from %s: %w", ComponentDescriptorFileName, err)
	}

	return NewComponentArchive(ctx, nil, cd, opts.ComponentFileSystem, closer), nil
}

func newComponentArchiveForFilesystem(ctx cpi.Context, cd *compdesc.ComponentDescriptor, opts ComponentArchiveOptions, closer ComponentCloser) (*ComponentArchive, error) {
	if cd == nil {
		cd = &compdesc.ComponentDescriptor{}
	}
	compdesc.DefaultComponent(cd)

	ca := NewComponentArchive(ctx, nil, cd, opts.ComponentFileSystem, closer)
	err := ca.writeCD()
	if err != nil {
		return nil, err
	}
	return ca, err
}
