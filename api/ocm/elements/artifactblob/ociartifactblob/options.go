package ociartifactblob

import (
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	base "ocm.software/ocm/api/utils/blobaccess/ociartifact"
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

// //////////////////////////////////////////////////////////////////////////////
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

// //////////////////////////////////////////////////////////////////////////////
// Blob (OCIArtifact) Options

func WithContext(ctx oci.ContextProvider) Option {
	return OptionFunc(func(opts *Options) {
		base.WithContext(ctx).ApplyTo(&opts.Blob)
	})
}

func WithVersion(v string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithVersion(v).ApplyTo(&opts.Blob)
	})
}

func WithPrinter(v common.Printer) Option {
	return OptionFunc(func(opts *Options) {
		base.WithPrinter(v).ApplyTo(&opts.Blob)
	})
}
