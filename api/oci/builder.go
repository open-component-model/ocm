package oci

import (
	"context"

	"github.com/open-component-model/ocm/api/credentials"
	"github.com/open-component-model/ocm/api/datacontext"
	"github.com/open-component-model/ocm/api/oci/internal"
)

func WithContext(ctx context.Context) internal.Builder {
	return internal.Builder{}.WithContext(ctx)
}

func WithCredentials(ctx credentials.Context) internal.Builder {
	return internal.Builder{}.WithCredentials(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) internal.Builder {
	return internal.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithRepositorySpecHandlers(reg RepositorySpecHandlers) internal.Builder {
	return internal.Builder{}.WithRepositorySpecHandlers(reg)
}

func New(mode ...datacontext.BuilderMode) Context {
	return internal.Builder{}.New(mode...)
}
