package blobhandler

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type (
	HandlerConfig   = cpi.BlobHandlerConfig
	HandlerOption   = cpi.BlobHandlerOption
	HandlerOptions  = cpi.BlobHandlerOptions
	HandlerRegistry = cpi.BlobHandlerRegistry
	HandlerKey      = cpi.BlobHandlerKey
)

func For(ctx cpi.ContextProvider) cpi.BlobHandlerRegistry {
	return ctx.OCMContext().BlobHandlers()
}
