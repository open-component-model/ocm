package utils

import (
	"fmt"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"golang.org/x/mod/modfile"
	"reflect"
	"runtime"
	"strings"
)

func GetRelativePathToProjectRoot(i ...int) (string, error) {
	iterations := general.OptionalDefaulted(20, i...)

	path := "."
	for count := 0; count < iterations; count++ {
		if ok, err := vfs.FileExists(osfs.OsFs, filepath.Join(path, "go.mod")); err != nil || ok {
			if err != nil {
				return "", fmt.Errorf("failed to check if go.mod exists: %v", err)
			}
			return path, nil
		}
		if count == iterations {
			return "", fmt.Errorf("could not find go.mod (within %d steps)", iterations)
		}
		path = filepath.Join(path, "..")
	}
	return "", nil
}

func GetModuleName() (string, error) {
	pathToRoot, err := GetRelativePathToProjectRoot()
	if err != nil {
		return "", err
	}
	pathToGoMod := filepath.Join(pathToRoot, "go.mod")
	// Read the content of the go.mod file
	data, err := vfs.ReadFile(osfs.OsFs, pathToGoMod)
	if err != nil {
		return "", err
	}

	// Parse the go.mod file
	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("error parsing go.mod file: %w", err)
	}

	// Print the module path
	return modFile.Module.Mod.Path, nil
}

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
	mod, err := GetModuleName()
	if err != nil {
		return "", err
	}
	path, ok := strings.CutPrefix(pkg, mod)
	if !ok {
		return "", fmt.Errorf("prefix %q not found in %q", mod, pkg)
	}
	return path, nil
}
