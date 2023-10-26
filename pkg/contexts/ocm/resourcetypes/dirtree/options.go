// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dirtree

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	base "github.com/open-component-model/ocm/pkg/common/accessio/blobaccess/dirtree"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/rpi"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	rpi.Options
	Blob base.Options
}

var _ rpi.GeneralOptionsProvider = (*Options)(nil)

func (o *Options) Apply(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	o.Blob.ApplyTo(&opts.Blob)
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

func mapBaseOption(opts *Options) *base.Options {
	return &opts.Blob
}

func wrapBase(o base.Option) Option {
	return optionutils.OptionWrapperFunc[*base.Options, *Options](o, mapBaseOption)
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return wrapBase(base.WithFileSystem(fs))
}

func WithExcludeFiles(files []string) Option {
	return wrapBase(base.WithExcludeFiles(files))
}

func WithIncludeFiles(files []string) Option {
	return wrapBase(base.WithIncludeFiles(files))
}

func WithFollowSymlinks(b ...bool) Option {
	return wrapBase(base.WithFollowSymlinks(b...))
}

func WithPreserveDir(b ...bool) Option {
	return wrapBase(base.WithPreserveDir(b...))
}

func WithCompressWithGzip(b ...bool) Option {
	return wrapBase(base.WithCompressWithGzip(b...))
}
