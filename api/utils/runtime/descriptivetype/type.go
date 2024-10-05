package descriptivetype

import (
	"ocm.software/ocm/api/utils/runtime"
)

// TypedObjectType is the appropriately extended type interface
// based on runtime.VersionTypedObjectType providing support for a functional and
// format description.
type TypedObjectType[T runtime.TypedObject] interface {
	runtime.TypedObjectType[T]
	TypeInfo
}

////////////////////////////////////////////////////////////////////////////////

type TypedObjectTypeObject[T runtime.TypedObject] struct {
	runtime.TypedObjectType[T]
	typeInfoImpl
	validator func(T) error
}

func NewTypedObjectTypeObject[E runtime.VersionedTypedObject](vt runtime.TypedObjectType[E], opts ...Option) *TypedObjectTypeObject[E] {
	target := &TypedObjectTypeObject[E]{
		TypedObjectType: vt,
	}
	t := NewOptionTargetWrapper(target, &target.typeInfoImpl)
	for _, o := range opts {
		o.ApplyTo(t)
	}
	return target
}

func (t *TypedObjectTypeObject[T]) Validate(e T) error {
	if t.validator == nil {
		return nil
	}
	return t.validator(e)
}
