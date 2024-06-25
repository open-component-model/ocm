package clictx

import (
	"github.com/open-component-model/ocm/api/clictx/internal"
)

type (
	Context = internal.Context
	OCI     = internal.OCI
	OCM     = internal.OCM
)

func DefaultContext() Context {
	return internal.DefaultContext
}
