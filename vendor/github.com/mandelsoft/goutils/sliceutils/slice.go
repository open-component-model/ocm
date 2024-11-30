package sliceutils

import (
	"slices"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/matcher"
)

// CopyAppend returns a new slice containing the additional elements appended to
// to the original slice.
func CopyAppend[E any](slice []E, elems ...E) []E {
	return append(slices.Clone(slice), elems...)
}

// AppendUnique appends elements to a slice, if they are not net contained.
func AppendUnique[S ~[]E, E comparable](in S, add ...E) S {
	for _, v := range add {
		if !slices.Contains(in, v) {
			in = append(in, v)
		}
	}
	return in
}

// CopyAppendUnique returns a new slice with additional elements appended,
// if they are not net contained.
func CopyAppendUnique[S ~[]E, E comparable](in S, add ...E) S {
	in = slices.Clone(in)
	for _, v := range add {
		if !slices.Contains(in, v) {
			in = append(in, v)
		}
	}
	return in
}

// AppendUniqueFunc returns appends additional elements,
// if they are considered by the given function not to be yet present.
func AppendUniqueFunc[S ~[]E, E any](in S, eq general.EqualsFunc[E], add ...E) S {
	for _, v := range add {
		if !slices.ContainsFunc(in, func(e E) bool { return eq(v, e) }) {
			in = append(in, v)
		}
	}
	return in
}

// CopyAppendUniqueFunc returns a new slice with additional elements appended,
// if they are considered by the given function not to be yet present.
func CopyAppendUniqueFunc[S ~[]E, E any](in S, eq general.EqualsFunc[E], add ...E) S {
	in = slices.Clone(in)
	for _, v := range add {
		if !slices.ContainsFunc(in, func(e E) bool { return eq(v, e) }) {
			in = append(in, v)
		}
	}
	return in
}

// Convert converts a slice to a slice with a more general element type.
func Convert[T, S any](a []S) []T {
	if a == nil {
		return nil
	}
	if generics.TypeOf[S]() == generics.TypeOf[T]() {
		return generics.Cast[[]T](a)
	}
	r := make([]T, len(a), len(a))
	for i, e := range a {
		r[i] = generics.Cast[T](e)
	}
	return r
}

// AsAny converts any slice to an interface slice.
func AsAny[S ~[]T, T any](s S) []any {
	return Convert[any](s)
}

// ConvertPointer converts a slice of pointers to
// an interface slice avoiding typed nil interfaces.
func ConvertPointer[T any, S ~[]P, E any, P generics.PointerType[E]](s S) []T {
	var _nil T

	// Preserve nil in case it matters.
	if s == nil {
		return nil
	}
	t := make([]T, len(s))
	for i, e := range s {
		if e == nil {
			t[i] = _nil
		} else {
			t[i] = generics.Cast[T](e)
		}
	}
	return t
}

// ConvertWith converts the element type of a slice
// using a converter function.
// Unfortunately this cannot be expressed in a type-safe way in Go.
// I MUST follow the type constraint I super S, which cannot be expressed in Go.
// If I == S the Transform function should be used, instead.
func ConvertWith[S, T, I any](in []S, c func(I) T) []T {
	if in == nil {
		return nil
	}
	r := make([]T, len(in))
	for i := range in {
		var s any = in[i]
		r[i] = c(generics.Cast[I](s))
	}
	return r
}

// Filter filters a slice by a matcher.Matcher.
func Filter[S ~[]E, E any](in S, f matcher.Matcher[E]) S {
	var r S
	for _, v := range in {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func Transform[S ~[]E, E any, T any](in S, m func(E) T) []T {
	r := make([]T, len(in))
	for i, v := range in {
		r[i] = m(v)
	}
	return r
}

func Reverse[S ~[]E, E any](in S) S {
	// if non-nil, provide always a separately modifiable copy.
	if in == nil {
		return in
	}
	s := slices.Clone(in)
	slices.Reverse(s)
	return s
}

// FilterType filters elements of a dedicated super type from
// a list of specialized types.
func FilterType[T any, S ~[]E, E any](elems S) []T {
	var r []T
	for _, e := range elems {
		if t, ok := generics.TryCast[T](e); ok {
			r = append(r, t)
		}
	}
	return r
}

// InitialSliceFor provides a new initial slice with length and capacity
// taken from the given one.
func InitialSliceFor[S ~[]E, E any](in S) S {
	return make(S, len(in), len(in))
}

// InitialSliceWithTypeFor is like InitialSliceFor, but provides a slice of the
// explicitly given type TS instead of the one from the given slice S.
func InitialSliceWithTypeFor[TS ~[]TE, TE any, S ~[]E, E any](in S) TS {
	return make(TS, len(in), len(in))
}

// AsSlice provides a slice for a given list of elements.
// If the elements are not og the same type, but only share a common super type,
// the intended super type must be passed as type parameter.
func AsSlice[T any](elems ...T) []T {
	return elems
}
