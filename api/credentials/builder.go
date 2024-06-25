package credentials

import (
	"context"

	"github.com/open-component-model/ocm/api/config"
	"github.com/open-component-model/ocm/api/credentials/internal"
	"github.com/open-component-model/ocm/api/datacontext"
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
