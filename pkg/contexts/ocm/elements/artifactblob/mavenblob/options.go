package mavenblob

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	base "github.com/open-component-model/ocm/pkg/blobaccess/maven"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
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

func WithCachingContext(cachectx datacontext.Context) Option {
	return wrapBase(base.WithCachingContext(cachectx))
}

func WithCachingFileSystem(fs vfs.FileSystem) Option {
	return wrapBase(base.WithCachingFileSystem(fs))
}

func WithCachingPath(p string) Option {
	return wrapBase(base.WithCachingPath(p))
}

func WithCredentials(c credentials.Credentials) Option {
	return wrapBase(base.WithCredentials(c))
}

func WithClassifier(c string) Option {
	return wrapBase(base.WithClassifier(c))
}

func WithOptionalClassifier(c *string) Option {
	return wrapBase(base.WithOptionalClassifier(c))
}

func WithExtension(e string) Option {
	return wrapBase(base.WithExtension(e))
}

func WithOptionalExtension(e *string) Option {
	return wrapBase(base.WithOptionalExtension(e))
}
