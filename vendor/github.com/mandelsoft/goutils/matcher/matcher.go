package matcher

import (
	"slices"
)

type Matcher[E any] func(E) bool

func NotInitial[E comparable](e E) bool {
	var _ini E
	return e != _ini
}

func Initial[E comparable](e E) bool {
	var _ini E
	return e == _ini
}

func Contains[E comparable](in ...E) func(E) bool {
	return func(e E) bool {
		v := slices.Contains(in, e)
		return v
	}
}

func ContainsFunc[E any](cmp func(E, E) int, in ...E) Matcher[E] {
	return func(e E) bool {
		return slices.ContainsFunc(in, func(c E) bool { return cmp(e, c) == 0 })
	}
}

func Equals[E comparable](e E) Matcher[E] {
	return func(c E) bool { return e == c }
}

func Not[E any](m Matcher[E]) Matcher[E] {
	return func(e E) bool { return !m(e) }
}

func And[E any](and ...Matcher[E]) Matcher[E] {
	return func(e E) bool {
		for _, m := range and {
			if !m(e) {
				return false
			}
		}
		return true
	}
}

func Or[E any](and ...Matcher[E]) Matcher[E] {
	return func(e E) bool {
		for _, m := range and {
			if m(e) {
				return true
			}
		}
		return false
	}
}
