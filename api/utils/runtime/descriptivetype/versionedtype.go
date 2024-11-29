package descriptivetype

import (
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

// VersionedTypedObjectType is the appropriately extended type interface
// based on runtime.VersionTypedObjectType providing support for a functional and
// format description.
type VersionedTypedObjectType[T runtime.VersionedTypedObject] interface {
	runtime.VersionedTypedObjectType[T]
	TypeInfo
}

////////////////////////////////////////////////////////////////////////////////

// VersionedTypeScheme is the appropriately extended scheme interface based on
// runtime.TypeScheme. Based on the additional type info a complete
// scheme description can be created calling the Describe method.
type VersionedTypeScheme[T runtime.VersionedTypedObject, R VersionedTypedObjectType[T]] interface {
	TypeScheme[T, R]
}

func MustNewDefaultVersionedTypeScheme[T runtime.VersionedTypedObject, R VersionedTypedObjectType[T], S VersionedTypeScheme[T, R]](name string, extender DescriptionExtender[R], unknown runtime.Unstructured, acceptUnknown bool, defaultdecoder runtime.TypedObjectDecoder[T], base ...VersionedTypeScheme[T, R]) VersionedTypeScheme[T, R] {
	scheme := runtime.MustNewDefaultTypeScheme[T, R](unknown, acceptUnknown, defaultdecoder, utils.Optional(base...))
	return &typeScheme[T, R, S]{
		name:        name,
		extender:    extender,
		_typeScheme: scheme,
		versioned:   true,
	}
}

// NewVersionedTypeScheme provides an TypeScheme implementation based on the interfaces
// and the default runtime.TypeScheme implementation.
func NewVersionedTypeScheme[T runtime.VersionedTypedObject, R VersionedTypedObjectType[T], S VersionedTypeScheme[T, R]](name string, extender DescriptionExtender[R], unknown runtime.Unstructured, acceptUnknown bool, base ...S) VersionedTypeScheme[T, R] {
	scheme := runtime.MustNewDefaultTypeScheme[T, R](unknown, acceptUnknown, nil, utils.Optional(base...))
	return &typeScheme[T, R, S]{
		name:        name,
		extender:    extender,
		_typeScheme: scheme,
		versioned:   true,
	}
}

////////////////////////////////////////////////////////////////////////////////

type VersionedTypedObjectTypeObject[T runtime.VersionedTypedObject] struct {
	runtime.VersionedTypedObjectType[T]
	typeInfoImpl
	validator func(T) error
}

func NewVersionedTypedObjectTypeObject[E runtime.VersionedTypedObject](vt runtime.VersionedTypedObjectType[E], opts ...Option) *VersionedTypedObjectTypeObject[E] {
	target := &VersionedTypedObjectTypeObject[E]{
		VersionedTypedObjectType: vt,
	}
	t := NewOptionTargetWrapper(target, &target.typeInfoImpl)
	for _, o := range opts {
		o.ApplyTo(t)
	}
	return target
}

func (t *VersionedTypedObjectTypeObject[T]) Validate(e T) error {
	if t.validator == nil {
		return nil
	}
	return t.validator(e)
}
