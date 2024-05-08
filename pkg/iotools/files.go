package iotools

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/utils"
)

func ListFiles(path string, fss ...vfs.FileSystem) ([]string, error) {
	var result []string
	fs := utils.FileSystem(fss...)
	err := vfs.Walk(fs, path, func(path string, info vfs.FileInfo, err error) error {
		result = append(result, path)
		return nil
	})
	return result, err
}
