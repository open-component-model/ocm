package npm

import (
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/tech/npm"
	"ocm.software/ocm/api/tech/npm/identity"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/stdopts"
)

type Option interface {
	ApplyTo(opts *Options)
}

type OptionFunc func(opts *Options)

func (f OptionFunc) ApplyTo(opts *Options) {
	f(opts)
}

type Options struct {
	stdopts.StandardContexts
	stdopts.PathFileSystem
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(
		o.LoggingContext.Value,
		o.CredentialContext.Value,
		o.CachingContext.Value,
	).Logger(npm.REALM).WithValues(keyValuePairs...)
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

// //////////////////////////////////////////////////////////////////////////////
// Option constructors

func WithCredentialContext(ctx credentials.ContextProvider) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetCredentialContext(ctx.CredentialsContext())
	})
}

func WithLoggingContext(ctx logging.ContextProvider) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetLoggingContext(ctx.LoggingContext())
	})
}

func WithCachingContext(ctx datacontext.Context) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetCachingContext(ctx)
	})
}

func WithCachingFileSystem(fs vfs.FileSystem) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetCachingFileSystem(fs)
	})
}

func WithCachingPath(p string) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetCachingPath(p)
	})
}

func WithCredentials(c credentials.Credentials) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetCredentials(c)
	})
}

func WithPathFileSystem(fs vfs.FileSystem) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetPathFileSystem(fs)
	})
}

// //////////////////////////////////////////////////////////////////////////////
// DataContext integration

func (o *Options) SetDataContext(ctx datacontext.Context) {
	if c, ok := ctx.(credentials.ContextProvider); ok {
		o.SetCredentialContext(c.CredentialsContext())
	}
	o.SetPathFileSystem(vfsattr.Get(ctx.AttributesContext()))
	o.SetCachingContext(ctx.AttributesContext())
}

var _ stdopts.DataContextOptionBag = (*Options)(nil)

func WithDataContext(ctx datacontext.Context) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetDataContext(ctx)
	})
}
