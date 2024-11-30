package generics

import (
	"reflect"
)

type PointerType[P any] interface {
	*P
}

func Pointer[T any](t T) *T {
	return &t
}

func TypeOf[T any]() reflect.Type {
	var t T
	return reflect.TypeOf(&t).Elem()
}
