package clictx

import (
	"ocm.software/ocm/api/cli/internal"
)

type (
	Context = internal.Context
	OCI     = internal.OCI
	OCM     = internal.OCM
)

func DefaultContext() Context {
	return internal.DefaultContext
}
