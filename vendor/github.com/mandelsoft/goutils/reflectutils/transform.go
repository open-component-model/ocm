package reflectutils

import (
	"reflect"
)

func MapValueToInterface(v reflect.Value) interface{} {
	return v.Interface()
}
