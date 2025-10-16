package git

import (
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/tmpcache"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/tech/git"
	"ocm.software/ocm/api/tech/git/identity"
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
	git.ClientOptions

	stdopts.StandardContexts
	stdopts.PathFileSystem
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(o.LoggingContext.Value, o.CredentialContext.Value).
		Logger(git.REALM).
		WithValues(keyValuePairs...)
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
	if o.Ref != "" {
		opts.Ref = o.Ref
	}
	if o.Commit != "" {
		opts.Commit = o.Commit
	}
}

func (o *Options) ConfigureAuthMethod() error {
	if o.ClientOptions.AuthMethod != nil {
		return nil
	}

	var err error
	if o.Credentials.Value != nil {
		if o.ClientOptions.AuthMethod, err = git.AuthFromCredentials(o.Credentials.Value); err != nil {
			return err
		}
	}

	if o.CredentialContext.Value == nil {
		return nil
	}

	creds, err := identity.GetCredentials(o.CredentialContext.Value, o.URL)
	if err != nil {
		return err
	}

	if creds != nil {
		if o.ClientOptions.AuthMethod, err = git.AuthFromCredentials(creds); err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) CachingFilesystem() (vfs.FileSystem, func() error, error) {
	if o.PathFileSystem.Value != nil {
		return o.PathFileSystem.Value, nil, nil
	}
	if o.CachingFileSystem.Value != nil {
		return o.CachingFileSystem.Value, nil, nil
	}

	if o.CachingContext.Value != nil {
		if fs := vfsattr.Get(o.CachingContext.Value); fs != nil {
			return fs, nil, nil
		}
		if fromtmp := tmpcache.Get(o.CachingContext.Value); fromtmp != nil {
			fs, err := projectionfs.New(fromtmp.Filesystem, fromtmp.Path)
			if err != nil {
				return nil, nil, err
			}
			return fs, nil, nil
		}
	}

	tmpfs, err := osfs.NewTempFileSystem()
	return tmpfs, func() error { return vfs.Cleanup(tmpfs) }, err
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

func WithPathFileSystem(fs vfs.FileSystem) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetPathFileSystem(fs)
	})
}

func WithCredentials(c credentials.Credentials) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetCredentials(c)
	})
}

// //////////////////////////////////////////////////////////////////////////////
// URL

type URLOptionBag interface {
	SetURL(v string)
}

func (o *Options) SetURL(v string) {
	o.URL = v
}

func WithURL(url string) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetURL(url)
	})
}

// //////////////////////////////////////////////////////////////////////////////
// Ref

type RefOptionBag interface {
	SetRef(v string)
}

func (o *Options) SetRef(v string) {
	o.Ref = v
}

func WithRef(ref string) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetRef(ref)
	})
}

// //////////////////////////////////////////////////////////////////////////////
// Commit

type CommitOptionBag interface {
	SetCommit(v string)
}

func (o *Options) SetCommit(v string) {
	o.Commit = v
}

func WithCommit(commit string) Option {
	return OptionFunc(func(opts *Options) {
		opts.SetCommit(commit)
	})
}
