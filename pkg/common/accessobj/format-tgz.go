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
	"compress/gzip"
	"io"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var FormatTGZ = TGZHandler{}

func init() {
	RegisterFormat(FormatTGZ)
}

type TGZHandler struct{}

// ApplyOption applies the configured path filesystem.
func (o TGZHandler) ApplyOption(options *Options) {
	f := o.Format()
	options.FileFormat = &f
}

func (_ TGZHandler) Format() accessio.FileFormat {
	return accessio.FormatTGZ
}

func (c TGZHandler) Open(info *AccessObjectInfo, acc AccessMode, path string, opts Options) (*AccessObject, error) {
	return DefaultOpenOptsFileHandling("tgz archive", info, acc, path, opts, c)
}

func (c TGZHandler) Create(info *AccessObjectInfo, path string, opts Options, mode vfs.FileMode) (*AccessObject, error) {
	return DefaultCreateOptsFileHandling("tgz archive", info, path, opts, mode, c)
}

// Write tars the current object and its artifacts.
func (c TGZHandler) Write(obj *AccessObject, path string, opts Options, mode vfs.FileMode) error {
	writer, err := opts.WriterFor(path, mode)
	if err != nil {
		return err
	}
	return c.WriteToStream(obj, writer, opts)
}

func (c TGZHandler) WriteToStream(obj *AccessObject, writer io.Writer, opts Options) error {
	gw := gzip.NewWriter(writer)
	if err := FormatTAR.WriteToStream(obj, gw, opts); err != nil {
		return err
	}
	return gw.Close()
}

// NewFromReader creates a new manifest builder from a input reader.
func (c TGZHandler) NewFromReader(info *AccessObjectInfo, acc AccessMode, in io.Reader, opts Options, closer Closer) (*AccessObject, error) {
	// the archive is untared to a memory fs that the builder can work
	// as it would be a default filesystem.

	in, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	return TarHandler{}.NewFromReader(info, acc, in, opts, closer)
}
