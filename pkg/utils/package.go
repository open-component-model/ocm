package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

const MODULE_PATH = "github.com/open-component-model/ocm"

func GetPackageNameForFunc(i interface{}) (string, error) {
	// Get the function's pointer
	ptr := reflect.ValueOf(i).Pointer()
	// Retrieve the function's runtime information
	funcForPC := runtime.FuncForPC(ptr)
	if funcForPC == nil {
		return "", fmt.Errorf("could not determine package name")
	}
	// Get the full name of the function, including the package path
	fullFuncName := funcForPC.Name()

	// Split the name to extract the package path
	// Assuming the format: "package/path.functionName"
	lastSlashIndex := strings.LastIndex(fullFuncName, "/")
	if lastSlashIndex == -1 {
		return "", fmt.Errorf("could not determine package name")
	}

	packagePath := fullFuncName[:lastSlashIndex]
	return packagePath, nil
}

func GetPackagePathFromProjectRootForFunc(i interface{}) (string, error) {
	pkg, err := GetPackageNameForFunc(i)
	if err != nil {
		return "", err
	}
	path, ok := strings.CutPrefix(pkg, "github.com/open-component-model/ocm/")
	if !ok {
		return "", fmt.Errorf("prefix %q not found in %q", MODULE_PATH, pkg)
	}
	return path, nil
}
