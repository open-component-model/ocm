package flagsetscheme

import (
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/runtime/descriptivetype"
)

// TypedObjectType is the appropriately extended type interface
// based on runtime.TypedObjectType.
type TypedObjectType[T runtime.TypedObject] interface {
	descriptivetype.TypedObjectType[T]

	ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler
}

////////////////////////////////////////////////////////////////////////////////

type TypedObjectTypeObject[E runtime.VersionedTypedObject] struct {
	*descriptivetype.TypedObjectTypeObject[E]
	typeInfoImpl
	validator func(E) error
}

func NewTypedObjectTypeObject[E runtime.VersionedTypedObject](vt runtime.TypedObjectType[E], opts ...TypeOption) *TypedObjectTypeObject[E] {
	target := &TypedObjectTypeObject[E]{
		TypedObjectTypeObject: descriptivetype.NewTypedObjectTypeObject[E](vt, optionutils.FilterMappedOptions[descriptivetype.OptionTarget](opts...)...),
	}
	t := NewOptionTargetWrapper[*TypedObjectTypeObject[E]](target, &target.typeInfoImpl)
	optionutils.ApplyOptions[OptionTarget](t, opts...)
	return t.target
}

func (t *TypedObjectTypeObject[E]) Validate(e E) error {
	if t.validator == nil {
		return nil
	}
	return t.validator(e)
}
