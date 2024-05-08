package standard

import (
	"reflect"
	"strings"

	"github.com/open-component-model/ocm/pkg/generics"
)

// HandleOption handles an option value (O) to be transferred to an
// options object (t) using the setter method provided by interface (I).
// To be transferred the options object must implement interface I
// and the option value must not be the zero value.
// The function handles teo cases:
//   - a pointer variable to the option value (transferred if pointer is not nil)
//   - a value variable (transferred if not zero).
func HandleOption[I any, O any](v O, t any) {
	vv := reflect.ValueOf(v)
	if !vv.IsZero() { // pointer of value variable))
		if opts, ok := t.(I); ok {
			vt := generics.TypeOf[O]()
			if vv.Kind() == reflect.Pointer {
				// switch to value behind the pointer (not nil)
				vv = vv.Elem()
				vt = vt.Elem()
			}
			oty := reflect.TypeOf(opts)
			ty := generics.TypeOf[I]()
			for i := 0; i < ty.NumMethod(); i++ {
				m := ty.Method(i)
				if strings.HasPrefix(m.Name, "Set") && m.IsExported() && m.Type.NumIn() == 1 && m.Type.NumOut() == 0 {
					if m.Type.In(0) == vt {
						c, _ := oty.MethodByName(m.Name)
						c.Func.Call([]reflect.Value{reflect.ValueOf(t), vv})
						return
					}
				}
			}
		}
	}
}
