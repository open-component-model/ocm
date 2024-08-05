package internal

import (
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
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
	Context       = cpi.Context
	Specification = metav1.MergeAlgorithmSpecification
	Value         = runtime.RawValue
	Hint          string
)

func Register(h Handler) {
	DefaultRegistry.RegisterHandler(h)
}

func Assign(hint Hint, spec *Specification) {
	DefaultRegistry.AssignHandler(hint, spec)
}
