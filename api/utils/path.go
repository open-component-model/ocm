package utils

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

var _osfs = osfs.New()

func FileSystem(fss ...vfs.FileSystem) vfs.FileSystem {
	return DefaultedFileSystem(_osfs, fss...)
}

func DefaultedFileSystem(def vfs.FileSystem, fss ...vfs.FileSystem) vfs.FileSystem {
	return general.OptionalDefaulted(def, fss...)
}
