package oci

import (
	"context"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci/internal"
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
