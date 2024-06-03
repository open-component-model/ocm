package wget

import (
	"io"
	"net/http"

	"github.com/open-component-model/ocm/pkg/blobaccess/wget"
)

type (
	Options = wget.Options
	Option  = wget.Option
)

func WithMimeType(mime string) Option {
	return wget.WithMimeType(mime)
}

func WithHeader(h http.Header) Option {
	return wget.WithHeader(h)
}

func WithVerb(v string) Option {
	return wget.WithVerb(v)
}

func WithBody(v io.Reader) Option {
	return wget.WithBody(v)
}

func WithNoRedirect(r ...bool) Option {
	return wget.WithNoRedirect(r...)
}
