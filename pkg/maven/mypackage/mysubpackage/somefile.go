// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package mysubpackage

import (
	"reflect"
	"runtime"
	"strings"
)

// Sample function to analyze
func SampleFunction() {}

func GetPackageName(i interface{}) string {
	// Get the function's pointer
	ptr := reflect.ValueOf(i).Pointer()
	// Retrieve the function's runtime information
	funcForPC := runtime.FuncForPC(ptr)
	if funcForPC == nil {
		return "unknown"
	}
	// Get the full name of the function, including the package path
	fullFuncName := funcForPC.Name()

	// Split the name to extract the package path
	// Assuming the format: "package/path.functionName"
	lastSlashIndex := strings.LastIndex(fullFuncName, "/")
	if lastSlashIndex == -1 {
		return "unknown"
	}

	packagePath := fullFuncName[:lastSlashIndex]
	return packagePath
}
