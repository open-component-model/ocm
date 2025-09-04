package wgetblob

import (
	"io"
	"net/http"

	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/wget"
	base "ocm.software/ocm/api/utils/blobaccess/wget"
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
// Local Options

func WithCredentialContext(credctx credentials.ContextProvider) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCredentialContext(credctx).ApplyTo(&opts.Blob)
	})
}

func WithLoggingContext(logctx logging.ContextProvider) Option {
	return OptionFunc(func(opts *Options) {
		base.WithLoggingContext(logctx).ApplyTo(&opts.Blob)
	})
}

func WithMimeType(mime string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithMimeType(mime).ApplyTo(&opts.Blob)
	})
}

func WithCredentials(creds credentials.Credentials) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCredentials(creds).ApplyTo(&opts.Blob)
	})
}

func WithHeader(h http.Header) Option {
	return OptionFunc(func(opts *Options) {
		base.WithHeader(h).ApplyTo(&opts.Blob)
	})
}

func WithVerb(v string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithVerb(v).ApplyTo(&opts.Blob)
	})
}

func WithBody(v io.Reader) Option {
	return OptionFunc(func(opts *Options) {
		base.WithBody(v).ApplyTo(&opts.Blob)
	})
}

func WithNoRedirect(r ...bool) Option {
	return OptionFunc(func(opts *Options) {
		wget.WithNoRedirect(r...).ApplyTo(&opts.Blob)
	})
}
