package stdopts

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"
)

type genricoption[S any, B any, T any] struct {
	value T
}

func (o *genricoption[S, B, T]) ApplyTo(opts B) {
	t := generics.TypeOf[S]()
	if t.NumMethod() != 1 {
		panic(fmt.Sprintf("invalid setter type %s", t))
	}
	m := t.Method(0)
	reflect.ValueOf(opts).MethodByName(m.Name).Call([]reflect.Value{reflect.ValueOf(o.value)})
}

func WithOption[S, B any, T any](v T) optionutils.Option[B] {
	var b B

	if _, ok := generics.TryCast[S](b); !ok {
		panic(fmt.Sprintf("%T must be %s", b, generics.TypeOf[S]()))
	}
	return &genricoption[S, B, T]{v}
}
