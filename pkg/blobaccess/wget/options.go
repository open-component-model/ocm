package wget

import (
	"github.com/mandelsoft/logging"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/wget/identity"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	CredentialContext credentials.Context

	LoggingContext logging.Context
	// MimeType defines the media type of the downloaded content
	MimeType string

	Credentials credentials.Credentials
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(o.LoggingContext, o.CredentialContext).Logger(REALM).WithValues(keyValuePairs)
}

func (o *Options) GetCredentials(url string) (credentials.Credentials, error) {
	switch {
	case o.Credentials != nil:
		return o.Credentials, nil
	case o.CredentialContext != nil:
		creds, err := credentials.CredentialsForConsumer(o.CredentialContext, identity.GetConsumerId(url), identity.IdentityMatcher)
		if err != nil {
			return nil, err
		}
		return creds, nil
	default:
		return nil, nil
	}
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.MimeType != "" {
		opts.MimeType = o.MimeType
	}
	if o.CredentialContext != nil {
		opts.CredentialContext = o.CredentialContext
	}
	if o.LoggingContext != nil {
		opts.LoggingContext = o.LoggingContext
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

type mimeType string

func (o mimeType) ApplyTo(opts *Options) {
	opts.MimeType = string(o)
}

func WithMimeType(mime string) Option {
	return mimeType(mime)
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
