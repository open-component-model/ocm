package blobhandler

import (
	"ocm.software/ocm/api/ocm/cpi"
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
