// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"os"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

// IsExecutable returns true if a given file is executable.
func IsExecutable(path string, fss ...vfs.FileSystem) bool {
	stat, err := accessio.FileSystem(fss...).Stat(path)
	if err != nil {
		return false
	}
	mode := stat.Mode()
	if !mode.IsRegular() {
		return false
	}
	if (mode & 0111) == 0 {
		return false
	}
	return true
}

// SplitPathList splits a path list.
// This is based on genSplit from strings/strings.go
func SplitPathList(pathList string) []string {
	if pathList == "" {
		return nil
	}
	n := 1
	for i := 0; i < len(pathList); i++ {
		if pathList[i] == os.PathListSeparator {
			n++
		}
	}
	start := 0
	a := make([]string, n)
	na := 0
	for i := 0; i+1 <= len(pathList) && na+1 < n; i++ {
		if pathList[i] == os.PathListSeparator {
			a[na] = pathList[start:i]
			na++
			start = i + 1
		}
	}
	a[na] = pathList[start:]
	return a[:na+1]
}
