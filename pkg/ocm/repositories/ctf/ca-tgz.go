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
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var CA_TGZ = ComponentTGZ{}

func init() {
	RegisterComponentArchiveFormat(CA_TGZ)
}

type ComponentTGZ struct{}

// ApplyOption applies the configured path filesystem.
func (c ComponentTGZ) ApplyOption(options *ComponentArchiveOptions) {
	options.ArchiveFormat = c
}

func (_ ComponentTGZ) String() string {
	return "tgz"
}

func (c ComponentTGZ) Open(ctx core.Context, path string, opts ComponentArchiveOptions) (*ComponentArchive, error) {
	// we expect that the path point to a tar
	file, err := opts.PathFileSystem.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open tgz archive from %s: %w", path, err)
	}
	defer file.Close()
	return newComponentArchiveFromTGZReader(ctx, file, opts, ComponentCloserFunction(func(ca *ComponentArchive) error { return c.close(ca, ComponentInfo{path, opts}) }))
}

func (c ComponentTGZ) Create(ctx core.Context, path string, cd *compdesc.ComponentDescriptor, opts ComponentArchiveOptions) (*ComponentArchive, error) {
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
func (c ComponentTGZ) Write(ca *ComponentArchive, path string, opts ComponentArchiveOptions) error {
	writer, err := opts.PathFileSystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	return c.WriteToStream(ca, writer, opts)
}

func (c ComponentTGZ) WriteToStream(ca *ComponentArchive, writer io.Writer, opts ComponentArchiveOptions) error {
	gw := gzip.NewWriter(writer)
	if err := CA_TAR.WriteToStream(ca, gw, opts); err != nil {
		return err
	}
	return gw.Close()
}

func (c ComponentTGZ) close(ca *ComponentArchive, info ComponentInfo) error {
	return c.Write(ca, info.Path, info.Options)
}

// newComponentArchiveFromTarReader creates a new manifest builder from a input reader.
func newComponentArchiveFromTGZReader(ctx core.Context, in io.Reader, opts ComponentArchiveOptions, closer ComponentCloser) (*ComponentArchive, error) {
	// the archive is untared to a memory fs that the builder can work
	// as it would be a default filesystem.

	in, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	return newComponentArchiveFromTarReader(ctx, in, opts, closer)
}
