package stringutils

import (
	"strings"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/sliceutils"
)

type Stringable interface {
	String() string
}

func AsString[T Stringable](s T) string {
	return s.String()
}

func CompareStringable[T Stringable](a, b T) int {
	return strings.Compare(a.String(), b.String())
}

func AsStringSlice[S ~[]T, T Stringable](s S) []string {
	return sliceutils.Transform(s, AsString[T])
}

func Join[L ~[]S, S Stringable](list L, seps ...string) string {
	separator := general.OptionalDefaulted(", ", seps...)
	sep := ""
	r := ""
	for _, e := range list {
		r += sep + e.String()
		sep = separator
	}
	return r
}

func JoinFunc[L ~[]E, E any](list L, separator string, f func(E) string) string {
	sep := ""
	r := ""
	for _, e := range list {
		r += sep + f(e)
		sep = separator
	}
	return r
}
