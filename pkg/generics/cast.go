// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package generics

func As[T any](o interface{}) T {
	var _nil T
	if o == nil {
		return _nil
	}
	return o.(T)
}

func AsE[T any](o interface{}, err error) (T, error) {
	var _nil T
	if o == nil {
		return _nil, err
	}
	return o.(T), err
}

// CastPointer casts a pointer/error result to an interface/error
// result.
// In Go this cannot be done directly, because returning a nil pinter
// for an interface return type, would result is a typed nil value for
// the interface, and not nil, if the pointer is nil.
// Unfortunately, the relation of the pointer (even the fact, that a pointer is
// expected)to the interface (even the fact, that an interface is expected)
// cannot be expressed with Go generics.
func CastPointer[T any](p any, err error) (T, error) {
	var _nil T
	if p == nil {
		return _nil, err
	}
	return p.(T), err
}
