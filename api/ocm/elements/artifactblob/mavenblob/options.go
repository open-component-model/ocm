package mavenblob

import (
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
	"ocm.software/ocm/api/tech/maven"
	base "ocm.software/ocm/api/utils/blobaccess/maven"
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

func WithHintForCoords(coords *maven.Coordinates) Option {
	if coords.IsPackage() {
		return WithHint(coords.GAV())
	}
	return nil
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return OptionFunc(func(opts *Options) {
		api.WithGlobalAccess(a).ApplyTo(&opts.Options)
	})
}

////////////////////////////////////////////////////////////////////////////////
// Local (Blob) Options

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

func WithCachingContext(cachectx datacontext.Context) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCachingContext(cachectx).ApplyTo(&opts.Blob)
	})
}

func WithCachingFileSystem(fs vfs.FileSystem) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCachingFileSystem(fs).ApplyTo(&opts.Blob)
	})
}

func WithCachingPath(p string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCachingPath(p).ApplyTo(&opts.Blob)
	})
}

func WithCredentials(c credentials.Credentials) Option {
	return OptionFunc(func(opts *Options) {
		base.WithCredentials(c).ApplyTo(&opts.Blob)
	})
}

func WithClassifier(c string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithClassifier(c).ApplyTo(&opts.Blob)
	})
}

func WithOptionalClassifier(c *string) Option {
	if c != nil {
		return WithClassifier(*c)
	}
	return nil
}

func WithExtension(e string) Option {
	return OptionFunc(func(opts *Options) {
		base.WithExtension(e).ApplyTo(&opts.Blob)
	})
}

func WithOptionalExtension(e *string) Option {
	if e != nil {
		return WithExtension(*e)
	}
	return nil
}
