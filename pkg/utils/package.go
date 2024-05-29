package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"golang.org/x/mod/modfile"
)

const GO_MOD = "go.mod"

// GetRelativePathToProjectRoot calculates the relative path to a go projects root directory.
// It therefore assumes that the project root is the directory containing the go.mod file.
// The optional parameter i determines how many directories the function will step up through, attempting to find a
// go.mod file. If it cannot find a directory with a go.mod file within i iterations, the function throws an error.
func GetRelativePathToProjectRoot(i ...int) (string, error) {
	iterations := general.OptionalDefaulted(20, i...)

	path := "."
	for count := 0; count < iterations; count++ {
		if ok, err := vfs.FileExists(osfs.OsFs, filepath.Join(path, GO_MOD)); err != nil || ok {
			if err != nil {
				return "", fmt.Errorf("failed to check if %s exists: %w", GO_MOD, err)
			}
			return path, nil
		}
		if count == iterations {
			return "", fmt.Errorf("could not find %s (within %d steps)", GO_MOD, iterations)
		}
		path = filepath.Join(path, "..")
	}
	return "", nil
}

// GetModuleName returns a go modules module name by finding and parsing the go.mod file.
func GetModuleName() (string, error) {
	pathToRoot, err := GetRelativePathToProjectRoot()
	if err != nil {
		return "", err
	}
	pathToGoMod := filepath.Join(pathToRoot, GO_MOD)
	// Read the content of the go.mod file
	data, err := vfs.ReadFile(osfs.OsFs, pathToGoMod)
	if err != nil {
		return "", err
	}

	// Parse the go.mod file
	modFile, err := modfile.Parse(GO_MOD, data, nil)
	if err != nil {
		return "", fmt.Errorf("error parsing %s file: %w", GO_MOD, err)
	}

	// Print the module path
	return modFile.Module.Mod.Path, nil
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

// GetPackageNameForCaller returns the package name of the function calling this function.
// The parameter offset is useful if this function is used within another utility function. With an offset of 1, you
// could specify that it should not return the package name of the immediate caller (the other utility function), but
// the package name of the function calling the utility function.
func GetPackageNameForCaller(offset ...int) (string, error) {
	pc, _, _, ok := runtime.Caller(general.OptionalDefaulted(0, offset...) + 1)
	if !ok {
		panic("unable to find caller")
	}
	return getPackageNameForFuncPC(pc)
}

// GetPackageNameForFunc provides the package name of a function. The function has to be passed as parameter. If no
// parameter is passed, the GetPackageNameForFunc will default to returning the package name of the calling function.
func GetPackageNameForFunc(f ...interface{}) (string, error) {
	// if no function is passed, assume the package name should be determined for the caller of this function
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("unable to find caller")
	}

	// Get the function's pointer
	var ptr uintptr
	if len(f) > 0 {
		ptr = reflect.ValueOf(general.Optional(f...)).Pointer()
	} else {
		ptr = pc
	}

	return getPackageNameForFuncPC(ptr)
}

func GetPackagePathFromProjectRootForFunc(f ...interface{}) (string, error) {
	var pkg string
	var err error

	if len(f) > 0 {
		pkg, err = GetPackageNameForFunc(f)
		if err != nil {
			return "", err
		}
	} else {
		pkg, err = GetPackageNameForCaller(1)
	}

	mod, err := GetModuleName()
	if err != nil {
		return "", err
	}
	path, ok := strings.CutPrefix(pkg, mod+"/")
	if !ok {
		return "", fmt.Errorf("prefix %q not found in %q", mod, pkg)
	}
	return path, nil
}
