package testutils

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"golang.org/x/mod/modfile"

	"github.com/open-component-model/ocm/pkg/utils/pkgutils"
)

const GO_MOD = "go.mod"

func GetPackagePathFromProjectRoot(i ...interface{}) (string, error) {
	pkg, err := pkgutils.GetPackageName(i...)
	if err != nil {
		return "", err
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
