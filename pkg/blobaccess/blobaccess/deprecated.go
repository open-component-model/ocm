package blobaccess

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/blobaccess/file"
)

// ForTemporaryFile wraps a temporary file into a BlobAccess, which does not need a close.
// Deprecated: ForTemporaryFile.
func ForTemporaryFileWithMeta(mime string, digest digest.Digest, size int64, temp vfs.File, fss ...vfs.FileSystem) bpi.BlobAccess {
	return file.BlobAccessForTemporaryFile(mime, temp, file.WithFileSystem(fss...), file.WithDigest(digest), file.WithSize(size))
}

// ForTemporaryFile wraps a temporary file into a BlobAccess, which does not need a close.
// Deprecated: ForTemporaryFilePath.
func ForTemporaryFilePathWithMeta(mime string, digest digest.Digest, size int64, temp string, fss ...vfs.FileSystem) BlobAccess {
	return file.BlobAccessForTemporaryFilePath(mime, temp, file.WithFileSystem(fss...), file.WithDigest(digest), file.WithSize(size))
}
