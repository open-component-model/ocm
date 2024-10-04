package git

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/tech/git"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/stdopts"
)

type Option = optionutils.Option[*Options]

type Options struct {
	git.ClientOptions

	stdopts.StandardContexts
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(o.LoggingContext.Value, o.CredentialContext.Value).Logger(git.REALM).WithValues(keyValuePairs...)
}

func (o *Options) Cache() *tmpcache.Attribute {
	if o.CachingPath.Value != "" {
		return tmpcache.New(o.CachingPath.Value, o.CachingFileSystem.Value)
	}
	if o.CachingContext.Value != nil {
		return tmpcache.Get(o.CachingContext.Value)
	}
	return tmpcache.Get(o.CredentialContext.Value)
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.CredentialContext.Value != nil {
		opts.CredentialContext = o.CredentialContext
	}
	if o.Credentials.Value != nil {
		opts.Credentials = o.Credentials
	}
	if o.LoggingContext.Value != nil {
		opts.LoggingContext = o.LoggingContext
	}
	if o.CachingFileSystem.Value != nil {
		opts.CachingFileSystem = o.CachingFileSystem
	}
	if o.URL != "" {
		opts.URL = o.URL
	}
	if o.Author.Name != "" && o.Author.Email != "" {
		opts.Author = o.Author
	}
	if o.Ref != "" {
		opts.Ref = o.Ref
	}
}

func option[S any, T any](v T) optionutils.Option[*Options] {
	return optionutils.WithGenericOption[S, *Options](v)
}

func WithCredentialContext(ctx credentials.ContextProvider) Option {
	return option[stdopts.CredentialContextOptionBag](ctx)
}

func WithLoggingContext(ctx logging.ContextProvider) Option {
	return option[stdopts.LoggingContextOptionBag](ctx)
}

func WithCachingContext(ctx datacontext.Context) Option {
	return option[stdopts.CachingContextOptionBag](ctx)
}

func WithCachingFileSystem(fs vfs.FileSystem) Option {
	return option[stdopts.CachingFileSystemOptionBag](fs)
}

func WithCachingPath(p string) Option {
	return option[stdopts.CachingPathOptionBag](p)
}

func WithCredentials(c credentials.Credentials) Option {
	return option[stdopts.CredentialsOptionBag](c)
}

////////////////////////////////////////////////////////////////////////////////

type URLOptionBag interface {
	SetURL(v string)
}

func (o *Options) SetURL(v string) {
	o.URL = v
}

func WithURL(url string) Option {
	return option[URLOptionBag](url)
}

////////////////////////////////////////////////////////////////////////////////

type RefOptionBag interface {
	SetRef(v string)
}

func (o *Options) SetRef(v string) {
	o.Ref = v
}

func WithRef(ref string) Option {
	return option[RefOptionBag](ref)
}
