package mavenblob

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	"ocm.software/ocm/api/tech/maven"
	base "ocm.software/ocm/api/utils/blobaccess/maven"
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

func WithHintForCoords(coords *maven.Coordinates) Option {
	if coords.IsPackage() {
		return WithHint(coords.GAV())
	}
	return optionutils.NoOption[*Options]{}
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
