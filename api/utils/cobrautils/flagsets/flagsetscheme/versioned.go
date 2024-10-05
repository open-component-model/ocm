package flagsetscheme

import (
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/runtime/descriptivetype"
)

// VersionTypedObjectType is the appropriately extended type interface
// based on runtime.VersionTypedObjectType.
type VersionTypedObjectType[T runtime.VersionedTypedObject] interface {
	descriptivetype.VersionedTypedObjectType[T]

	ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler
}

////////////////////////////////////////////////////////////////////////////////

type VersionedTypedObjectTypeObject[E runtime.VersionedTypedObject] struct {
	*descriptivetype.VersionedTypedObjectTypeObject[E]
	typeInfoImpl
	validator func(E) error
}

var _ TypeInfo = (*VersionedTypedObjectTypeObject[runtime.VersionedTypedObject])(nil)

func NewVersionedTypedObjectTypeObject[E runtime.VersionedTypedObject](vt runtime.VersionedTypedObjectType[E], opts ...TypeOption) *VersionedTypedObjectTypeObject[E] {
	target := &VersionedTypedObjectTypeObject[E]{
		VersionedTypedObjectTypeObject: descriptivetype.NewVersionedTypedObjectTypeObject[E](vt, optionutils.FilterMappedOptions[descriptivetype.OptionTarget](opts...)...),
	}
	t := NewOptionTargetWrapper[*VersionedTypedObjectTypeObject[E]](target, &target.typeInfoImpl)
	optionutils.ApplyOptions[OptionTarget](t, opts...)
	return t.target
}

func (t *VersionedTypedObjectTypeObject[E]) Validate(e E) error {
	if t.validator == nil {
		return nil
	}
	return t.validator(e)
}
