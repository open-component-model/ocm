// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package generics

func As[T any](o interface{}) T {
	var zero T
	if o == nil {
		return zero
	}
	return o.(T)
}

func AsE[T any](o interface{}, err error) (T, error) {
	var zero T
	if o == nil {
		return zero, err
	}
	return o.(T), err
}

func CastSlice[S, T any](in ...S) []T {
	r := make([]T, len(in))
	for i := range in {
		var s any = in[i]
		r[i] = s.(T)
	}
	return r
}
