package helmblob

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	base "ocm.software/ocm/api/utils/blobaccess/helm"
	common "ocm.software/ocm/api/utils/misc"
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
// Blob (Helm) Options

func WithFileSystem(fs vfs.FileSystem) Option {
	return OptionFunc(func(opts *Options) {
		base.WithFileSystem(fs).ApplyTo(&opts.Blob)
	})
}

func WithContext(ctx oci.ContextProvider) Option {
	return OptionFunc(func(opts *Options) {
		base.WithContext(ctx).ApplyTo(&opts.Blob)
	})
}

func WithIVersion(v string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithVersion(v).ApplyTo(&opts.Blob)
	})
}

func WithIVersionOverride(v string, flag ...bool) Option {
	return OptionFunc(func(opts *Options) {
		base.WithVersionOverride(v, flag...).ApplyTo(&opts.Blob)
	})
}

func WithCACert(v string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCACert(v).ApplyTo(&opts.Blob)
	})
}

func WithCACertFile(v string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCACertFile(v).ApplyTo(&opts.Blob)
	})
}

func WithHelmRepository(v string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithHelmRepository(v).ApplyTo(&opts.Blob)
	})
}

func WithPrinter(v common.Printer) Option {
	return OptionFunc(func(opts *Options) {
		base.WithPrinter(v).ApplyTo(&opts.Blob)
	})
}
