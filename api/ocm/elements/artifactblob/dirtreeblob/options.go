package dirtreeblob

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	base "ocm.software/ocm/api/utils/blobaccess/dirtree"
)

type Option interface {
	ApplyTo(opts *Options)
}

type OptionFunc func(opts *Options)

func (f OptionFunc) ApplyTo(opts *Options) {
	f(opts)
}

type Options struct {
	api.Options
	Blob base.Options
}

var (
	_ api.GeneralOptionsProvider = (*Options)(nil)
	_ Option                     = (*Options)(nil)
)

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	o.Options.ApplyTo(&opts.Options)
	o.Blob.ApplyTo(&opts.Blob)
}

func (o *Options) Apply(opts ...Option) {
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(o)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return OptionFunc(func(opts *Options) {
		api.WithHint(h).ApplyTo(&opts.Options)
	})
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return OptionFunc(func(opts *Options) {
		api.WithGlobalAccess(a).ApplyTo(&opts.Options)
	})
}

////////////////////////////////////////////////////////////////////////////////
// DirTree BlobAccess Options

func WithFileSystem(fs vfs.FileSystem) Option {
	return OptionFunc(func(opts *Options) {
		base.WithFileSystem(fs).ApplyTo(&opts.Blob)
	})
}

func WithExcludeFiles(files []string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithExcludeFiles(files).ApplyTo(&opts.Blob)
	})
}

func WithIncludeFiles(files []string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithIncludeFiles(files).ApplyTo(&opts.Blob)
	})
}

func WithFollowSymlinks(b ...bool) Option {
	return OptionFunc(func(opts *Options) {
		base.WithFollowSymlinks(b...).ApplyTo(&opts.Blob)
	})
}

func WithPreserveDir(b ...bool) Option {
	return OptionFunc(func(opts *Options) {
		base.WithPreserveDir(b...).ApplyTo(&opts.Blob)
	})
}

func WithCompressWithGzip(b ...bool) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCompressWithGzip(b...).ApplyTo(&opts.Blob)
	})
}
