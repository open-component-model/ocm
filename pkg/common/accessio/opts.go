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
	"io"
	"os"

	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/open-component-model/ocm/pkg/common/compression"
	"github.com/open-component-model/ocm/pkg/errors"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

type Options struct {
	// FilePath is the path of the repository base in the filesystem
	FileFormat *FileFormat `json:"fileFormat"`
	// FileSystem is the virtual filesystem to evaluate the file path. Default is the OS filesytem
	// or the filesystem defined as base filesystem for the context
	// This configuration option is not available for the textual representation of
	// the repository specification
	PathFileSystem vfs.FileSystem `json:"-"`
	// Representation is the virtual filesystem to represent the active repository cache.
	// This configuration option is not available for the textual representation of
	// the repository specification
	Representation vfs.FileSystem `json:"-"`
	// File is an opened file object to use instead of the path and path filesystem
	// It should never be closed if given to support temporary files
	File vfs.File `json:"-"`
	// Reader provides a one time access to the content (archive xontent only)
	// The resulting access is therefore temporarily and cannot be written back
	// to its origin, but to other destinations.
	Reader io.ReadCloser `json:"-"`
}

var _ Option = &Options{}

var _osfs = osfs.New()

func (o Options) ApplyOption(options *Options) {
	if o.PathFileSystem != nil {
		options.PathFileSystem = o.PathFileSystem
	}
	if o.Representation != nil {
		options.Representation = o.Representation
	}
	if o.FileFormat != nil {
		options.FileFormat = o.FileFormat
	}
	if o.File != nil {
		options.File = o.File
	}
	if o.Reader != nil {
		options.Reader = o.Reader
	}
}

func (o Options) Default() Options {
	if o.PathFileSystem == nil {
		o.PathFileSystem = _osfs
	}
	return o
}

func (o Options) DefaultFormat(fmt FileFormat) Options {
	if o.FileFormat == nil {
		o.FileFormat = &fmt
	}
	return o
}

func (o Options) ValidForPath(path string) error {
	count := 0
	if path != "" {
		count++
	}
	if o.File != nil {
		count++
	}
	if o.Reader != nil {
		count++
	}
	if count > 1 {
		return errors.ErrInvalid("only path,, file or reader can be set")
	}
	return nil
}

func (o Options) DefaultForPath(path string) (Options, error) {
	if err := o.ValidForPath(path); err != nil {
		return o, err
	}
	if o.FileFormat == nil {
		var fmt *FileFormat
		var err error
		switch {
		case o.Reader != nil:
			r, _, err := compression.AutoDecompress(o.Reader)
			if err == nil {
				o.Reader = AddCloser(r, o.Reader)
				f := FormatTar
				fmt = &f
			}
		case o.File != nil:
			fmt, err = DetectFormatForFile(o.File)
		default:
			fmt, err = DetectFormat(path, o.PathFileSystem)
		}
		if err == nil {
			o.FileFormat = fmt
		}
		return o, err
	}
	return o, nil
}

func (o Options) WriterFor(path string, mode vfs.FileMode) (io.WriteCloser, error) {
	if err := o.ValidForPath(path); err != nil {
		return nil, err
	}
	var writer io.WriteCloser
	var err error
	if o.File == nil {
		writer, err = o.PathFileSystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode&0666)
	} else {
		writer = NopWriteCloser(o.File)
		err = o.File.Truncate(0)
	}
	return writer, err
}

// ApplyOptions applies the given list options on these options,
// and then returns itself (for convenient chaining).
func (o Options) ApplyOptions(opts ...Option) Options {
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyOption(&o)
		}
	}
	return o
}

// Option is the interface to specify different archive options
type Option interface {
	ApplyOption(options *Options)
}

// PathFileSystem set the evaluation filesystem for the path name
func PathFileSystem(fs vfs.FileSystem) Option {
	return opt_PFS{fs}
}

type opt_PFS struct {
	vfs.FileSystem
}

// ApplyOption applies the configured path filesystem.
func (o opt_PFS) ApplyOption(options *Options) {
	options.PathFileSystem = o.FileSystem
}

// RepresentationFileSystem set the evaltuation filesystem for the path name
func RepresentationFileSystem(fs vfs.FileSystem) Option {
	return opt_RFS{fs}
}

type opt_RFS struct {
	vfs.FileSystem
}

// ApplyOption applies the configured path filesystem.
func (o opt_RFS) ApplyOption(options *Options) {
	options.Representation = o.FileSystem
}

// File set open file to use
func File(file vfs.File) Option {
	return opt_F{file}
}

type opt_F struct {
	vfs.File
}

// ApplyOption applies the configured open file
func (o opt_F) ApplyOption(options *Options) {
	options.File = o.File
}

// Reader set open reader to use
func Reader(reader io.ReadCloser) Option {
	return opt_R{reader}
}

type opt_R struct {
	io.ReadCloser
}

// ApplyOption applies the configured open file
func (o opt_R) ApplyOption(options *Options) {
	options.Reader = o.ReadCloser
}

////////////////////////////////////////////////////////////////////////////////

func AccessOptions(opts ...Option) Options {
	return Options{}.ApplyOptions(opts...).Default()
}
