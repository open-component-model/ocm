package hpi

import (
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/valuemergehandler/internal"
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
			SetFor(octx, NewRegistry(internal.DefaultRegistry))
		case datacontext.MODE_CONFIGURED:
			SetFor(octx, internal.DefaultRegistry.Copy())
		case datacontext.MODE_INITIAL:
			SetFor(octx, NewRegistry())
		}
	}
}
