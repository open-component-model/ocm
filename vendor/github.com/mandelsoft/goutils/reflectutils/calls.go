package reflectutils

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/sliceutils"
)

// GetInterfaceMethod gets the method of an interface
// with one method.
func GetInterfaceMethod[M any]() reflect.Method {
	t := generics.TypeOf[M]()
	if t.NumMethod() != 1 {
		panic(fmt.Sprintf("invalid setter type %s", t))
	}
	return t.Method(0)
}

// CallMethodByInterfaceVA calls a void method on object o with
// one argument a. The method is specified by the interface
// M, which should implement exactly one appropriate method.
func CallMethodByInterfaceVA[M, B any](o B, a interface{}) {
	CallMethodByNameVA[B](GetInterfaceMethod[M]().Name, o, a)
}

func CallMethodByInterface[M, B any](o B, args ...interface{}) []interface{} {
	return CallMethodByName[B](GetInterfaceMethod[M]().Name, o, args...)
}

func CallMethodByNameVA[B any](n string, o B, a interface{}) {
	reflect.ValueOf(o).MethodByName(n).Call([]reflect.Value{reflect.ValueOf(a)})
}

func CallMethodByName[B any](n string, o B, args ...interface{}) []interface{} {
	v := sliceutils.Transform(args, reflect.ValueOf)
	r := reflect.ValueOf(o).MethodByName(n).Call(v)
	return sliceutils.Transform(r, MapValueToInterface)
}
