// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package wgetaccess

import (
	"io"
	"net/http"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/pkg/blobaccess/wget"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
)

type (
	Options = wget.Options
	Option  = wget.Option
)

func WithCredentialContext(ctx credentials.ContextProvider) Option {
	return wget.WithCredentialContext(ctx)
}

func WithLoggingContext(ctx logging.ContextProvider) Option {
	return wget.WithLoggingContext(ctx)
}

func WithMimeType(mime string) Option {
	return wget.WithMimeType(mime)
}

func WithCredentials(c credentials.Credentials) Option {
	return wget.WithCredentials(c)
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
