package general

import (
	"cmp"
)

type (
	CompareFunc[E any]  func(a, b E) int
	EqualsFunc[E any]   func(E, E) bool
	ContainsFunc[E any] func(E) bool
)

func EqualsForCompareFuncFor[E any](cmp CompareFunc[E]) EqualsFunc[E] {
	return func(a, b E) bool {
		return cmp(a, b) == 0
	}
}

func ContainsForCompareFuncFor[E any](a E, cmp CompareFunc[E]) ContainsFunc[E] {
	return func(b E) bool {
		return cmp(a, b) == 0
	}
}

func ContainsForEqualsFuncFor[E any](a E, eq EqualsFunc[E]) ContainsFunc[E] {
	return func(b E) bool {
		return eq(a, b)
	}
}

type Equals[E any] interface {
	Equals(a E) bool
}

func EqualsFuncFor[E Equals[E]]() EqualsFunc[E] {
	return func(a, b E) bool {
		return a.Equals(b)
	}
}

func ContainsFuncFor[E Equals[E]](a E) ContainsFunc[E] {
	return func(b E) bool {
		return a.Equals(b)
	}
}

func EqualsComparable[E comparable](a, b E) bool {
	return a == b
}

func CompareOrdered[E cmp.Ordered](a, b E) int {
	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}
