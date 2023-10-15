// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dirtree

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess/dirtree"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/rpi"
)

type Option = rpi.ResourceOption[*Options]

type Options struct {
	rpi.Options
	DirTree dirtree.Options
}

var _ rpi.GeneralOptionsProvider = (*Options)(nil)

func (o *Options) Apply(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	o.DirTree.ApplyToDirtreeOptions(&opts.DirTree)
}

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return rpi.WrapHint[Options](h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return rpi.WrapGlobalAccess[Options](a)
}

////////////////////////////////////////////////////////////////////////////////
// DirTree BlobAccess Options

type blobaccessoption struct {
	opt dirtree.Option
}

func (w blobaccessoption) ApplyTo(opts *Options) {
	w.opt.ApplyToDirtreeOptions(&opts.DirTree)
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return blobaccessoption{dirtree.WithFileSystem(fs)}
}

func WithExcludeFiles(files []string) Option {
	return blobaccessoption{dirtree.WithExcludeFiles(files)}
}

func WithIncludeFiles(files []string) Option {
	return blobaccessoption{dirtree.WithIncludeFiles(files)}
}

func WithFollowSymlinks(b ...bool) Option {
	return blobaccessoption{dirtree.WithFollowSymlinks(b...)}
}

func WithPreserveDir(b ...bool) Option {
	return blobaccessoption{dirtree.WithPreserveDir(b...)}
}

func WithCompressWithGzip(b ...bool) Option {
	return blobaccessoption{dirtree.WithCompressWithGzip(b...)}
}
