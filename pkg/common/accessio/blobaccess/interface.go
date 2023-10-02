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

// Validatable is an optional interface for DataAccess
// implementations or any other object, which might reach
// an error state. The error can then be queried with
// the method ErrorProvider.Validate.
// This is used to support objects with access methods not
// returning an error. If the object is not valid,
// those methods return an unknown/default state, but
// the object should be queryable for its state.
type Validatable = accessio.Validatable

// Validate checks whether a blob access
// is in error state. If yes, an appropriate
// error is returned.
func Validate(o BlobAccess) error {
	return accessio.ValidateObject(o)
}
