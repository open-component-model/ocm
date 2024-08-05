package config

import (
	"context"

	"ocm.software/ocm/api/config/internal"
	"ocm.software/ocm/api/datacontext"
)

func WithContext(ctx context.Context) internal.Builder {
	return internal.Builder{}.WithContext(ctx)
}

func WithSharedAttributes(ctx datacontext.AttributesContext) internal.Builder {
	return internal.Builder{}.WithSharedAttributes(ctx)
}

func WithConfigTypeScheme(scheme ConfigTypeScheme) internal.Builder {
	return internal.Builder{}.WithConfigTypeScheme(scheme)
}

func New(mode ...datacontext.BuilderMode) Context {
	return internal.Builder{}.New(mode...)
}
