package cpi

import (
	"context"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/internal"
)

func WithContext(ctx context.Context) internal.Builder {
	return internal.Builder{}.WithContext(ctx)
}

func WithCredentials(ctx credentials.Context) internal.Builder {
	return internal.Builder{}.WithCredentials(ctx)
}

func WithOCIRepositories(ctx oci.Context) internal.Builder {
	return internal.Builder{}.WithOCIRepositories(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) internal.Builder {
	return internal.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithRepositoryDelegation(reg RepositoryDelegationRegistry) internal.Builder {
	return internal.Builder{}.WithRepositoryDelegation(reg)
}

func WithAccessypeScheme(scheme AccessTypeScheme) internal.Builder {
	return internal.Builder{}.WithAccessTypeScheme(scheme)
}

func WithRepositorySpecHandlers(reg RepositorySpecHandlers) internal.Builder {
	return internal.Builder{}.WithRepositorySpecHandlers(reg)
}

func WithBlobHandlers(reg BlobHandlerRegistry) internal.Builder {
	return internal.Builder{}.WithBlobHandlers(reg)
}

func WithBlobDigesters(reg BlobDigesterRegistry) internal.Builder {
	return internal.Builder{}.WithBlobDigesters(reg)
}

func New(mode ...datacontext.BuilderMode) Context {
	return internal.Builder{}.New(mode...)
}
