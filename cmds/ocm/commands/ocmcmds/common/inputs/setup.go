package inputs

import (
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
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
			SetFor(octx, NewInputTypeScheme(nil, DefaultInputTypeScheme))
		case datacontext.MODE_CONFIGURED:
			SetFor(octx, DefaultInputTypeScheme.Copy())
		case datacontext.MODE_INITIAL:
			SetFor(octx, NewInputTypeScheme(nil))
		}
	}
}
