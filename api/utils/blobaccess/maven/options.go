package maven

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/tech/maven"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/stdopts"
)

type Option = optionutils.Option[*Options]

type Options struct {
	stdopts.StandardContexts

	maven.FileCoordinates
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(o.LoggingContext.Value, o.CredentialContext.Value).Logger(maven.REALM).WithValues(keyValuePairs...)
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

func (o *Options) GetCredentials(repo *maven.Repository, groupId string) (maven.Credentials, error) {
	if repo.IsFileSystem() {
		return nil, nil
	}

	switch {
	case o.Credentials.Value != nil:
		return MapCredentials(o.Credentials.Value), nil
	case o.CredentialContext.Value != nil:
		return GetCredentials(o.CredentialContext.Value, repo, groupId)
	default:
		return nil, nil
	}
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.CredentialContext.Value != nil {
		opts.CredentialContext = o.CredentialContext
	}
	if o.LoggingContext.Value != nil {
		opts.LoggingContext = o.LoggingContext
	}
	if o.CachingFileSystem.Value != nil {
		opts.CachingFileSystem = o.CachingFileSystem
	}
	if o.Credentials.Value != nil {
		opts.Credentials = o.Credentials
	}
	if o.Classifier != nil {
		opts.Classifier = o.Classifier
	}
	if o.Extension != nil {
		opts.Extension = o.Extension
	}
	if o.MediaType != nil {
		opts.MediaType = o.MediaType
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

type ClassifierOptionBag interface {
	SetClassifier(v string)
}

func (o *Options) SetClassifier(v string) {
	o.Classifier = &v
}

func WithClassifier(c string) Option {
	return option[ClassifierOptionBag](c)
}

func WithOptionalClassifier(c *string) Option {
	if c != nil {
		return WithClassifier(*c)
	}
	return &optionutils.NoOption[*Options]{}
}

////////////////////////////////////////////////////////////////////////////////

type MediaTypeOptionBag interface {
	SetMediaType(v string)
}

func (o *Options) SetMediaType(v string) {
	o.MediaType = &v
}

func WithMediaType(c string) Option {
	return option[MediaTypeOptionBag](c)
}

func WithOptionalMediaType(c *string) Option {
	if c != nil {
		return WithMediaType(*c)
	}
	return &optionutils.NoOption[*Options]{}
}

////////////////////////////////////////////////////////////////////////////////

type ExtensionOptionBag interface {
	SetExtension(v string)
}

func (o *Options) SetExtension(v string) {
	o.Extension = &v
}

func WithExtension(e string) Option {
	return option[ExtensionOptionBag](e)
}

func WithOptionalExtension(e *string) Option {
	if e != nil {
		return WithExtension(*e)
	}
	return &optionutils.NoOption[*Options]{}
}

////////////////////////////////////////////////////////////////////////////////

func (o *Options) SetDataContext(ctx datacontext.Context) {
	if c, ok := ctx.(credentials.ContextProvider); ok {
		o.SetCredentialContext(c.CredentialsContext())
	}
	o.SetCachingContext(ctx)
}

func WithDataContext(ctx datacontext.ContextProvider) Option {
	return option[stdopts.DataContextOptionBag](ctx)
}
