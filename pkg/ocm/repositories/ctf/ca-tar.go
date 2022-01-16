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
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var CA_TAR = ComponentTAR{}

func init() {
	RegisterComponentArchiveFormat(CA_TAR)
}

type ComponentTAR struct{}

// ApplyOption applies the configured path filesystem.
func (c ComponentTAR) ApplyOption(options *ComponentArchiveOptions) {
	options.ArchiveFormat = c
}

func (_ ComponentTAR) String() string {
	return "tar"
}

func (c ComponentTAR) Open(ctx core.Context, path string, opts ComponentArchiveOptions) (*ComponentArchive, error) {
	// we expect that the path point to a tar
	file, err := opts.PathFileSystem.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open tar archive from %s: %w", path, err)
	}
	defer file.Close()
	return newComponentArchiveFromTarReader(ctx, file, opts, ComponentCloserFunction(func(ca *ComponentArchive) error { return c.close(ca, ComponentInfo{path, opts}) }))
}

func (c ComponentTAR) Create(ctx core.Context, path string, cd *compdesc.ComponentDescriptor, opts ComponentArchiveOptions) (*ComponentArchive, error) {
	ok, err := vfs.Exists(opts.PathFileSystem, path)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, vfs.ErrExist
	}

	ca, err := newComponentArchiveForFilesystem(ctx, cd, opts, ComponentCloserFunction(func(ca *ComponentArchive) error { return c.close(ca, ComponentInfo{path, opts}) }))
	if err != nil {
		return nil, err
	}
	err = c.Write(ca, path, opts)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// Write tars the current components descriptor and its artifacts.
func (c ComponentTAR) Write(ca *ComponentArchive, path string, opts ComponentArchiveOptions) error {
	writer, err := opts.PathFileSystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	return c.WriteToStream(ca, writer, opts)
}

func (c ComponentTAR) WriteToStream(ca *ComponentArchive, writer io.Writer, opts ComponentArchiveOptions) error {
	tw := tar.NewWriter(writer)

	// write component descriptor
	cdBytes, err := compdesc.Encode(ca.ComponentDescriptor, opts.EncodeOptions)
	if err != nil {
		return fmt.Errorf("unable to encode component descriptor: %w", err)
	}
	cdHeader := &tar.Header{
		Name:    ComponentDescriptorFileName,
		Size:    int64(len(cdBytes)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(cdHeader); err != nil {
		return fmt.Errorf("unable to write component descriptor header: %w", err)
	}
	if _, err := io.Copy(tw, bytes.NewBuffer(cdBytes)); err != nil {
		return fmt.Errorf("unable to write component descriptor content: %w", err)
	}

	// add all blobs
	err = tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     BlobsDirectoryName,
		Mode:     0644,
		ModTime:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("unable to write blob directory: %w", err)
	}

	blobs, err := vfs.ReadDir(ca.fs, BlobsDirectoryName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to read blob directory: %w", err)
	}
	for _, blobInfo := range blobs {
		blobpath := BlobPath(blobInfo.Name())
		header := &tar.Header{
			Name:    blobpath,
			Size:    blobInfo.Size(),
			Mode:    0644,
			ModTime: time.Now(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("unable to write blob header: %w", err)
		}

		blob, err := ca.fs.Open(blobpath)
		if err != nil {
			return fmt.Errorf("unable to open blob: %w", err)
		}
		if _, err := io.Copy(tw, blob); err != nil {
			return fmt.Errorf("unable to write blob content: %w", err)
		}
		if err := blob.Close(); err != nil {
			return fmt.Errorf("unable to close blob %s: %w", blobpath, err)
		}
	}

	return tw.Close()
}

func (c ComponentTAR) close(ca *ComponentArchive, info ComponentInfo) error {
	return c.Write(ca, info.Path, info.Options)
}

// newComponentArchiveFromTarReader creates a new manifest builder from a input reader.
func newComponentArchiveFromTarReader(ctx core.Context, in io.Reader, opts ComponentArchiveOptions, closer ComponentCloser) (*ComponentArchive, error) {
	// the archive is untared to a memory fs that the builder can work
	// as it would be a default filesystem.

	if opts.ComponentFileSystem == nil {
		opts.ComponentFileSystem = memoryfs.New()
	}

	if err := ExtractTarToFs(opts.ComponentFileSystem, in); err != nil {
		return nil, fmt.Errorf("unable to extract tar: %w", err)
	}

	return newComponentArchiveFromFilesystem(ctx, opts, closer)
}
