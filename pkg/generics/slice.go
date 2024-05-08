package generics

import (
	"golang.org/x/exp/slices"
)

func AppendedSlice[E any](slice []E, elems ...E) []E {
	return append(slices.Clone(slice), elems...)
}
