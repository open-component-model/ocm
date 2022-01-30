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
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

type Options struct {
	// FilePath is the path of the repository base in the filesystem
	FileFormat *accessio.FileFormat `json:"fileFormat"`
	// FileSystem is the virtual filesystem to evaluate the file path. Default is the OS filesytem
	// or the filesystem defined as base filesystem for the context
	// This configuration option is not available for the textual representation of
	// the repository specification
	PathFileSystem vfs.FileSystem `json:"-"`
	// Representation is the virtual filesystem to represent the active repository cache.
	// This configuration option is not available for the textual representation of
	// the repository specification
	Representation vfs.FileSystem `json:"-"`
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
}

func (o Options) Default() Options {
	if o.PathFileSystem == nil {
		o.PathFileSystem = _osfs
	}
	if o.FileFormat == nil {
		fmt := accessio.FormatDirectory
		o.FileFormat = &fmt
	}
	return o
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

// PathFileSystem set the evaltuation filesystem for the path name
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

////////////////////////////////////////////////////////////////////////////////

func AccessOptions(opts ...Option) Options {
	return Options{}.ApplyOptions(opts...).Default()
}
