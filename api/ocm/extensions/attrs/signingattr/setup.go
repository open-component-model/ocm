package signingattr

import (
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/tech/signing"
)

func init() {
	datacontext.RegisterSetupHandler(datacontext.SetupHandlerFunction(setupContext))
}

func setupContext(mode datacontext.BuilderMode, ctx datacontext.Context) {
	if octx, ok := ctx.(Context); ok {
		switch mode {
		case datacontext.MODE_SHARED:
			fallthrough
		case datacontext.MODE_DEFAULTED:
			// do nothing, fallback to the default attribute lookup
		case datacontext.MODE_EXTENDED:
			Set(octx, signing.NewRegistry(signing.DefaultRegistry().HandlerRegistry(), signing.DefaultRegistry().KeyRegistry()))
		case datacontext.MODE_CONFIGURED:
			Set(octx, signing.DefaultRegistry().Copy())
		case datacontext.MODE_INITIAL:
			Set(octx, signing.NewRegistry(nil, nil))
		}
	}
}
