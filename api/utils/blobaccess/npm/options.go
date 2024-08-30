package npm

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/builtin/npm/identity"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/tech/npm"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/stdopts"
)

type Option = optionutils.Option[*Options]

type Options struct {
	stdopts.StandardContexts
	stdopts.PathFileSystem
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(o.LoggingContext.Value, o.CredentialContext.Value, o.CachingContext.Value).Logger(npm.REALM).WithValues(keyValuePairs...)
}

func (o *Options) FileSystem() vfs.FileSystem {
	if o.PathFileSystem.Value != nil {
		return o.PathFileSystem.Value
	}
	if o.CachingFileSystem.Value != nil {
		return o.CachingFileSystem.Value
	}
	if o.CachingContext.Value != nil {
		return vfsattr.Get(o.CachingContext.Value)
	}
	return osfs.OsFs
}

func (o *Options) GetCredentials(repo string, pkg string) (cpi.Credentials, error) {
	switch {
	case o.Credentials.Value != nil:
		return o.Credentials.Value, nil
	case o.CredentialContext.Value != nil:
		return identity.GetCredentials(o.CredentialContext.Value, repo, pkg)
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
	if o.PathFileSystem.Value != nil {
		opts.PathFileSystem = o.PathFileSystem
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

func WithPathFileSystem(fs vfs.FileSystem) Option {
	return option[stdopts.PathFileSystemOptionBag](fs)
}

func (o *Options) SetDataContext(ctx datacontext.Context) {
	if c, ok := ctx.(credentials.ContextProvider); ok {
		o.CredentialContext.Value = c.CredentialsContext()
	}
	o.PathFileSystem.Value = vfsattr.Get(ctx.AttributesContext())
	o.CachingContext.Value = ctx.AttributesContext()
}

var _ stdopts.DataContextOptionBag = (*Options)(nil)

func WithDataContext(ctx datacontext.Context) Option {
	return option[stdopts.DataContextOptionBag](ctx)
}
