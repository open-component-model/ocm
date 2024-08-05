// Package hpi contains the Handler Programming Interface for
// value merge handlers
package hpi

import (
	"ocm.software/ocm/api/datacontext"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/valuemergehandler/internal"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

// resolve package cycle among default merge handler and
// labelmergehandler by separating commonly used types
// into this package

// same problem for the embedding into the OCM environment
// required for the ocm.Context access.
// Because of this cycle, the registry implementation and the
// required types have to be placed into the internal package of
// ocm and forwarded to the cpi packages. From there it can be consumed
// here to break the dependency cycle.

type (
	Context       = internal.Context
	Handler       = internal.Handler
	Config        = internal.Config
	Registry      = internal.Registry
	Specification = internal.Specification
	Value         = internal.Value
	Hint          = internal.Hint
)

const KIND_VALUE_MERGE_ALGORITHM = metav1.KIND_VALUE_MERGE_ALGORITHM

func Register(h Handler) {
	internal.Register(h)
}

func Assign(hint Hint, spec *Specification) {
	internal.Assign(hint, spec)
}

func NewSpecification(algo string, cfg ...Config) (*Specification, error) {
	raw, err := runtime.AsRawMessage(utils.Optional(cfg...))
	if err != nil {
		return nil, err
	}
	return &Specification{
		Algorithm: algo,
		Config:    raw,
	}, nil
}

func NewRegistry(base ...Registry) Registry {
	return internal.NewRegistry(base...)
}

func LabelHint(name string, optversion ...string) Hint {
	hint := "label:" + name
	v := utils.Optional(optversion...)
	if v != "" {
		hint += "@" + v
	}
	return Hint(hint)
}

////////////////////////////////////////////////////////////////////////////////

const ATTR_MERGE_HANDLERS = "ocm.software/ocm/api/ocm/valuemergehandlers"

func For(ctx cpi.ContextProvider) Registry {
	if ctx == nil {
		return internal.DefaultRegistry
	}
	return ctx.OCMContext().GetAttributes().GetOrCreateAttribute(ATTR_MERGE_HANDLERS, create).(Registry)
}

func create(datacontext.Context) interface{} {
	return NewRegistry(internal.DefaultRegistry)
}

func SetFor(ctx datacontext.Context, registry Registry) {
	ctx.GetAttributes().SetAttribute(ATTR_MERGE_HANDLERS, registry)
}
