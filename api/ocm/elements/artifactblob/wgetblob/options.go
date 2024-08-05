package wgetblob

import (
	"io"
	"net/http"

	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/wget"
	base "ocm.software/ocm/api/utils/blobaccess/wget"
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
// Local Options

func mapBaseOption(opts *Options) *base.Options {
	return &opts.Blob
}

func wrapBase(o base.Option) Option {
	return optionutils.OptionWrapperFunc[*base.Options, *Options](o, mapBaseOption)
}

func WithCredentialContext(credctx credentials.ContextProvider) Option {
	return wrapBase(base.WithCredentialContext(credctx))
}

func WithLoggingContext(logctx logging.ContextProvider) Option {
	return wrapBase(base.WithLoggingContext(logctx))
}

func WithMimeType(mime string) Option {
	return wrapBase(base.WithMimeType(mime))
}

func WithCredentials(creds credentials.Credentials) Option {
	return wrapBase(base.WithCredentials(creds))
}

func WithHeader(h http.Header) Option {
	return wrapBase(base.WithHeader(h))
}

func WithVerb(v string) Option {
	return wrapBase(base.WithVerb(v))
}

func WithBody(v io.Reader) Option {
	return wrapBase(base.WithBody(v))
}

func WithNoRedirect(r ...bool) Option {
	return wrapBase(wget.WithNoRedirect(r...))
}
