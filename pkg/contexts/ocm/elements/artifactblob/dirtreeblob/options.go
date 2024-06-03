package dirtreeblob

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"

	base "github.com/open-component-model/ocm/pkg/blobaccess/dirtree"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactblob/api"
)

type Option = optionutils.Option[*Options]

type Options struct {
	api.Options
	Blob base.Options
}

var (
	_ api.GeneralOptionsProvider = (*Options)(nil)
	_ Option                     = (*Options)(nil)
)

func (o *Options) ApplyTo(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	o.Blob.ApplyTo(&opts.Blob)
}

func (o *Options) Apply(opts ...Option) {
	optionutils.ApplyOptions(o, opts...)
}

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return api.WrapHint[Options](h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return api.WrapGlobalAccess[Options](a)
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
