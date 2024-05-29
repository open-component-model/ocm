package pkgutils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// GetPackageName gets the package name for an object, a type, a function or a caller offset.
//
// Examples:
//
//	GetPackageName(1)
//	GetPackageName(&MyStruct{})
//	GetPackageName(GetPackageName)
//	GetPackageName(generics.TypeOf[MyStruct]())
func GetPackageName(i ...interface{}) (string, error) {
	if len(i) == 0 {
		i = []interface{}{0}
	}
	if t, ok := i[0].(reflect.Type); ok {
		pkgpath := t.PkgPath()
		if pkgpath == "" {
			return "", fmt.Errorf("unable to determine package name")
		}
		return pkgpath, nil
	}
	v := reflect.ValueOf(i[0])
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Func:
		return getPackageNameForFuncPC(v.Pointer())
	case reflect.Struct, reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		pkgpath := v.Type().PkgPath()
		if pkgpath == "" {
			return "", fmt.Errorf("unable to determine package name")
		}
		return pkgpath, nil
	default:
		offset, err := CastInt(v.Interface())
		if err != nil {
			return "", err
		}
		pc, _, _, ok := runtime.Caller(offset + 1)
		if !ok {
			return "", fmt.Errorf("unable to find caller")
		}
		return getPackageNameForFuncPC(pc)
	}
}

func getPackageNameForFuncPC(pc uintptr) (string, error) {
	// Retrieve the function's runtime information
	funcForPC := runtime.FuncForPC(pc)
	if funcForPC == nil {
		return "", fmt.Errorf("could not determine package name")
	}
	// Get the full name of the function, including the package path
	fullFuncName := funcForPC.Name()

	// Split the name to extract the package path
	// Assuming the format: "package/path.functionName"
	lastSlashIndex := strings.LastIndex(fullFuncName, "/")
	if lastSlashIndex == -1 {
		panic("unable to find package name")
	}

	funcIndex := strings.Index(fullFuncName[lastSlashIndex:], ".")
	packagePath := fullFuncName[:lastSlashIndex+funcIndex]

	return packagePath, nil
}

func CastInt(i interface{}) (int, error) {
	switch v := i.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unable to cast %T into int", i)
	}
}
