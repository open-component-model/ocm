package options

import (
	"reflect"

	"github.com/mandelsoft/goutils/generics"
)

func FilterOptions[T any, O any](opts []O) []T {
	var found []T

	t := generics.TypeOf[T]()
	for _, o := range opts {
		if reflect.TypeOf(o).AssignableTo(t) {
			found = append(found, generics.Cast[T](o))
		}
	}
	return found
}
