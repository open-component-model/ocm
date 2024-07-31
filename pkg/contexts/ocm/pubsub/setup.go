package pubsub

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

func init() {
	datacontext.RegisterSetupHandler(datacontext.SetupHandlerFunction(setupContext))
}

func setupContext(mode datacontext.BuilderMode, ctx datacontext.Context) {
	if octx, ok := ctx.(cpi.Context); ok {
		switch mode {
		case datacontext.MODE_SHARED:
			fallthrough
		case datacontext.MODE_DEFAULTED:
			// do nothing, fallback to the default attribute lookup
		case datacontext.MODE_EXTENDED:
			SetSchemeFor(octx, NewTypeScheme(DefaultTypeScheme))
			SetProvidersFor(octx, NewProviderRegistry(DefaultRegistry))
		case datacontext.MODE_CONFIGURED:
			s := NewTypeScheme(nil)
			s.AddKnownTypes(DefaultTypeScheme)
			SetSchemeFor(octx, s)
			r := NewProviderRegistry(nil)
			r.AddKnownProviders(DefaultRegistry)
			SetProvidersFor(octx, r)
		case datacontext.MODE_INITIAL:
			SetSchemeFor(octx, NewTypeScheme())
			SetProvidersFor(octx, NewProviderRegistry())
		}
	}
}
