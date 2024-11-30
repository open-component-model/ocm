package generics

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
)

// TryCast is like Cast, but reports
// whether the assertion is possible or not.
func TryCast[T any](o any) (T, bool) {
	var _nil T
	if o == nil {
		return _nil, true
	}
	var i any = o
	t, ok := i.(T)
	return t, ok
}

// TryCastE casts one type parameter to another type parameter,
// which have a subtype relation.
// This cannot be described by type parameter constraints in Go, because
// constraints may not be type parameters again.
func TryCastE[T any](o any) (T, error) {
	var _nil T
	if o == nil {
		return _nil, nil
	}
	var s any = o
	if t, ok := s.(T); ok {
		return t, nil
	}
	return _nil, errors.ErrInvalid("type", fmt.Sprintf("%T", o))
}

// Cast asserts a type given by a type parameter for a value
// This is not directly suppoerted by Go.
//
//	func [O any](...) {
//	   x := i.(O)
//	}
func Cast[T any](o any) T {
	var _nil T
	if o == nil {
		return _nil
	}
	var i any = o
	t := i.(T)
	return t
}

// CastR cast a result type to a dedicated Type T
// for a factory function with an additional error result
// Nil will be mapped to the initial value of the target type
func CastR[T any](o any, err error) (T, error) {
	var _nil T
	if o == nil {
		return _nil, err
	}

	var s any = o
	if t, ok := s.(T); ok {
		return t, err
	}
	return _nil, errors.ErrInvalid("type", fmt.Sprintf("%T", o))
}

// CastPointer maps a pointer P to an interface type I
// avoiding typed nil pointers. Nil pointers will be mapped
// to nil interfaces.
func CastPointer[I any, E any, P PointerType[E]](e P) I {
	var _nil I
	if e == nil {
		return _nil
	}
	var i any = e
	return i.(I)
}

// CastPointerR maps a pointer P to an interface type I
// for a factory function with an additional error result
// avoiding typed nil pointers. Nil pointers will be mapped
// to nil interfaces.
func CastPointerR[I any, E any, P PointerType[E]](e P, err error) (I, error) {
	var _nil I
	if e == nil || err != nil {
		return _nil, err
	}
	var i any = e
	return i.(I), nil
}
