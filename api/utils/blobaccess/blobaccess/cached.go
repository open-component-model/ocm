package blobaccess

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/file"
)

func ForCachedBlobAccess(blob BlobAccess, fss ...vfs.FileSystem) (BlobAccess, error) {
	fs := utils.FileSystem(fss...)

	r, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	f, err := vfs.TempFile(fs, "", "cachedBlob*")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		return nil, err
	}
	f.Close()

	return file.BlobAccessForTemporaryFilePath(blob.MimeType(), f.Name(), file.WithFileSystem(fs)), nil
}
