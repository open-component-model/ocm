package wget

import (
	"io"
	"net/http"

	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/logging"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/tech/wget/identity"
	"ocm.software/ocm/api/utils"
	ocmlog "ocm.software/ocm/api/utils/logging"
)

type Option = optionutils.Option[*Options]

type Options struct {
	CredentialContext credentials.Context
	LoggingContext    logging.Context
	// Header to be passed in the http request
	Header http.Header
	// Verb is the http verb to be used for the request
	Verb string
	// Body is the body to be included in the http request
	Body io.Reader
	// NoRedirect allows to disable redirects
	NoRedirect *bool
	// MimeType defines the media type of the downloaded content
	MimeType string
	// Credentials allows to pass credentials and certificates for the http communication
	Credentials credentials.Credentials
}

func (o *Options) Logger(keyValuePairs ...interface{}) logging.Logger {
	return ocmlog.LogContext(o.LoggingContext, o.CredentialContext).Logger(REALM).WithValues(keyValuePairs...)
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
	if o.Credentials != nil {
		opts.Credentials = o.Credentials
	}
	if o.Header != nil {
		opts.Header = o.Header
	}
	if o.Verb != "" {
		opts.Verb = o.Verb
	}
	if o.Body != nil {
		opts.Body = o.Body
	}
	if o.NoRedirect != nil {
		opts.NoRedirect = o.NoRedirect
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

////////////////////////////////////////////////////////////////////////////////

type header http.Header

func (o header) ApplyTo(opts *Options) {
	opts.Header = http.Header(o)
}

func WithHeader(h http.Header) Option {
	return header(h)
}

////////////////////////////////////////////////////////////////////////////////

type verb string

func (o verb) ApplyTo(opts *Options) {
	opts.Verb = string(o)
}

func WithVerb(v string) Option {
	return verb(v)
}

////////////////////////////////////////////////////////////////////////////////

type body struct {
	io.Reader
}

func (o *body) ApplyTo(opts *Options) {
	if o.Reader != nil {
		opts.Body = io.Reader(o)
	}
}

func WithBody(v io.Reader) Option {
	return &body{v}
}

////////////////////////////////////////////////////////////////////////////////

type noredirect bool

func (o noredirect) ApplyTo(opts *Options) {
	opts.NoRedirect = utils.BoolP(o)
}

func WithNoRedirect(r ...bool) Option {
	return noredirect(utils.OptionalDefaultedBool(true, r...))
}
