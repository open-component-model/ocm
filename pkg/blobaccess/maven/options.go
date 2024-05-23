// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package maven

import (
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/maven/identity"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/tmpcache"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/maven"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	CredentialContext credentials.Context
	LoggingContext    logging.Context
	CachingContext    datacontext.Context
	FileSystem        vfs.FileSystem
	CachingPath       string
	// Credentials allows to pass credentials and certificates for the http communication
	Credentials credentials.Credentials
	// Classifier defines the classifier of the maven file coordinates
	Classifier *string
	// Extension defines the extension of the maven file coordinates
	Extension *string
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(o.LoggingContext, o.CredentialContext).Logger(maven.REALM).WithValues(keyValuePairs...)
}

func (o *Options) Cache() *tmpcache.Attribute {
	if o.CachingPath != "" {
		return tmpcache.New(o.CachingPath, o.FileSystem)
	}
	return tmpcache.Get(o.CachingContext)
}

func (o *Options) GetCredentials(repoUrl, groupId string) (maven.Credentials, error) {
	switch {
	case o.Credentials != nil:
		return MapCredentials(o.Credentials), nil
	case o.CredentialContext != nil:
		consumerid, err := identity.GetConsumerId(repoUrl, groupId)
		if err != nil {
			return nil, err
		}
		creds, err := credentials.CredentialsForConsumer(o.CredentialContext, consumerid, identity.IdentityMatcher)
		if err != nil {
			return nil, err
		}
		return MapCredentials(creds), nil
	default:
		return nil, nil
	}
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.CredentialContext != nil {
		opts.CredentialContext = o.CredentialContext
	}
	if o.LoggingContext != nil {
		opts.LoggingContext = o.LoggingContext
	}
	if o.FileSystem != nil {
		opts.FileSystem = o.FileSystem
	}
	if o.Credentials != nil {
		opts.Credentials = o.Credentials
	}
	if o.Classifier != nil {
		opts.Classifier = o.Classifier
	}
	if o.Extension != nil {
		opts.Extension = o.Extension
	}
}

////////////////////////////////////////////////////////////////////////////////

type context struct {
	credentials.Context
}

func (o context) ApplyTo(opts *Options) {
	opts.CredentialContext = o
}

func WithCredentialContext(ctx credentials.ContextProvider) Option {
	return context{ctx.CredentialsContext()}
}

////////////////////////////////////////////////////////////////////////////////

type loggingContext struct {
	logging.Context
}

func (o loggingContext) ApplyTo(opts *Options) {
	opts.LoggingContext = o
}

func WithLoggingContext(ctx logging.ContextProvider) Option {
	return loggingContext{ctx.LoggingContext()}
}

////////////////////////////////////////////////////////////////////////////////

type fileSystem struct {
	fs vfs.FileSystem
}

func (o *fileSystem) ApplyTo(opts *Options) {
	opts.FileSystem = o.fs
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return &fileSystem{fs: fs}
}

////////////////////////////////////////////////////////////////////////////////

type creds struct {
	credentials.Credentials
}

func (o creds) ApplyTo(opts *Options) {
	opts.Credentials = o.Credentials
}

func WithCredentials(c credentials.Credentials) Option {
	return creds{c}
}

////////////////////////////////////////////////////////////////////////////////

type classifier string

func (o classifier) ApplyTo(opts *Options) {
	opts.Classifier = optionutils.PointerTo(string(o))
}

func WithClassifier(c string) Option {
	return classifier(c)
}

func WithOptionalClassifier(c *string) Option {
	if c != nil {
		return WithClassifier(*c)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type extension string

func (o extension) ApplyTo(opts *Options) {
	opts.Extension = optionutils.PointerTo(string(o))
}

func WithExtension(e string) Option {
	return extension(e)
}

func WithOptionalExtension(e *string) Option {
	if e != nil {
		return WithExtension(*e)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
