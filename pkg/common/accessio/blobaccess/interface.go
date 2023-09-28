package blobaccess

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess/dirtree"
)

type BlobAccess = accessio.BlobAccess

func ForString(mime string, data string) BlobAccess {
	return accessio.BlobAccessForString(mime, data)
}

func ForData(mime string, data []byte) BlobAccess {
	return accessio.BlobAccessForData(mime, data)
}

func ForFile(mime string, path string, fss ...vfs.FileSystem) BlobAccess {
	return accessio.BlobAccessForFile(mime, path, fss...)
}

func ForDirTree(path string, opts ...dirtree.Option) (BlobAccess, error) {
	return dirtree.BlobAccessForDirTree(path, opts...)
}
