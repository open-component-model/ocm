package pubsub

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

const ATTR_PUBSUB_TYPES = "github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub"

type Attribute struct {
	ProviderRegistry
	TypeScheme
}

func For(ctx cpi.ContextProvider) *Attribute {
	if ctx == nil {
		return &Attribute{
			ProviderRegistry: DefaultRegistry,
			TypeScheme:       DefaultTypeScheme,
		}
	}
	return ctx.OCMContext().GetAttributes().GetOrCreateAttribute(ATTR_PUBSUB_TYPES, create).(*Attribute)
}

func create(datacontext.Context) interface{} {
	return &Attribute{
		ProviderRegistry: NewProviderRegistry(DefaultRegistry),
		TypeScheme:       NewTypeScheme(DefaultTypeScheme),
	}
}

func SetSchemeFor(ctx cpi.ContextProvider, registry TypeScheme) {
	attr := For(ctx)
	attr.TypeScheme = registry
}

func SetProvidersFor(ctx cpi.ContextProvider, registry ProviderRegistry) {
	attr := For(ctx)
	attr.ProviderRegistry = registry
}
