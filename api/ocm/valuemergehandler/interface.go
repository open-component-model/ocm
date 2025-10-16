package valuemergehandler

import (
	_ "ocm.software/ocm/api/ocm/valuemergehandler/config"
	_ "ocm.software/ocm/api/ocm/valuemergehandler/handlers"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
	"ocm.software/ocm/api/ocm/valuemergehandler/internal"
)

type (
	Context       = internal.Context
	Handler       = internal.Handler
	Handlers      = internal.Handlers
	Config        = internal.Config
	Registry      = internal.Registry
	Specification = internal.Specification
	Value         = internal.Value
)

const (
	KIND_VALUE_MERGE_ALGORITHM = hpi.KIND_VALUE_MERGE_ALGORITHM
	KIND_VALUESET              = "value set"
)

func NewSpecification(algo string, cfg Config) (*Specification, error) {
	return hpi.NewSpecification(algo, cfg)
}

func NewRegistry(base ...Registry) Registry {
	return internal.NewRegistry(base...)
}

func For(ctx cpi.ContextProvider) Registry {
	return hpi.For(ctx)
}

func SetFor(ctx datacontext.Context, registry Registry) {
	hpi.SetFor(ctx, registry)
}
