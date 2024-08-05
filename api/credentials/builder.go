package credentials

import (
	"context"

	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials/internal"
	"ocm.software/ocm/api/datacontext"
)

func WithContext(ctx context.Context) internal.Builder {
	return internal.Builder{}.WithContext(ctx)
}

func WithConfigs(ctx config.Context) internal.Builder {
	return internal.Builder{}.WithConfig(ctx)
}

func WithRepositoyTypeScheme(scheme RepositoryTypeScheme) internal.Builder {
	return internal.Builder{}.WithRepositoyTypeScheme(scheme)
}

func WithStandardConumerMatchers(matchers internal.IdentityMatcherRegistry) internal.Builder {
	return internal.Builder{}.WithStandardConumerMatchers(matchers)
}

func New(mode ...datacontext.BuilderMode) Context {
	return internal.Builder{}.New(mode...)
}
