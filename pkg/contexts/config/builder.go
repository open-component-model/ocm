package config

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/config/internal"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
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
