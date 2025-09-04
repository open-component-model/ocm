package dockermultiblob

import (
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	base "ocm.software/ocm/api/utils/blobaccess/dockermulti"
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
// Docker Multi-Blob Options

func WithVariants(names ...string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithVariants(names...).ApplyTo(&opts.Blob)
	})
}

func WithVersion(v string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithVersion(v).ApplyTo(&opts.Blob)
	})
}

func WithOrigin(o common.NameVersion) Option {
	return OptionFunc(func(opts *Options) {
		base.WithOrigin(o).ApplyTo(&opts.Blob)
	})
}

func WithPrinter(p common.Printer) Option {
	return OptionFunc(func(opts *Options) {
		base.WithPrinter(p).ApplyTo(&opts.Blob)
	})
}
