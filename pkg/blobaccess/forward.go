package blobaccess

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/blobaccess/file"
	"github.com/open-component-model/ocm/pkg/blobaccess/standard"
)

///////////
// Standard
///////////

func DataAccessForReaderFunction(reader func() (io.ReadCloser, error), origin string) bpi.DataAccess {
	return standard.DataAccessForReaderFunction(reader, origin)
}

// DataAccessForBytes wraps a bytes slice into a DataAccess.
// Deprecated: used DataAccessForData.
func DataAccessForBytes(data []byte, origin ...string) bpi.DataSource {
	return standard.DataAccessForData(data, origin...)
}

// DataAccessForData wraps a bytes slice into a DataAccess.
func DataAccessForData(data []byte, origin ...string) bpi.DataSource {
	return standard.DataAccessForData(data, origin...)
}

func DataAccessForString(data string, origin ...string) bpi.DataSource {
	return standard.DataAccessForString(data, origin...)
}

// ForString wraps a string into a BlobAccess, which does not need a close.
func ForString(mime string, data string) BlobAccess {
	return standard.ForString(mime, data)
}

func ProviderForString(mime, data string) BlobAccessProvider {
	return standard.ProviderForString(mime, data)
}

// ForData wraps data into a BlobAccess, which does not need a close.
func ForData(mime string, data []byte) BlobAccess {
	return standard.ForData(mime, data)
}

func ProviderForData(mime string, data []byte) BlobAccessProvider {
	return standard.ProviderForData(mime, data)
}

///////////
// File
///////////

func DataAccessForFile(fs vfs.FileSystem, path string) DataAccess {
	return file.DataAccess(fs, path)
}

func ForFile(mime string, path string, fss ...vfs.FileSystem) bpi.BlobAccess {
	return file.BlobAccess(mime, path, fss...)
}

func ProviderForFile(mime string, path string, fss ...vfs.FileSystem) bpi.BlobAccessProvider {
	return file.ProviderForFile(mime, path, fss...)
}

func ForFileWithCloser(closer io.Closer, mime string, path string, fss ...vfs.FileSystem) bpi.BlobAccess {
	return file.BlobAccessWithCloser(closer, mime, path, fss...)
}

func ForTemporaryFile(mime string, temp vfs.File, fss ...vfs.FileSystem) bpi.BlobAccess {
	return file.ForTemporaryFile(mime, temp, fss...)
}

func ForTemporaryFileWithMeta(mime string, digest digest.Digest, size int64, temp vfs.File, fss ...vfs.FileSystem) bpi.BlobAccess {
	return file.ForTemporaryFileWithMeta(mime, digest, size, temp, fss...)
}

func ForTemporaryFilePath(mime string, temp string, fss ...vfs.FileSystem) BlobAccess {
	return file.ForTemporaryFilePath(mime, temp, fss...)
}

func ForTemporaryFilePathWithMeta(mime string, digest digest.Digest, size int64, temp string, fss ...vfs.FileSystem) BlobAccess {
	return file.ForTemporaryFilePathWithMeta(mime, digest, size, temp, fss...)
}

// TempFile holds a temporary file that should be kept open.
// Close should never be called directly.
// It can be passed to another responsibility realm by calling Release
// For example to be transformed into a TemporaryBlobAccess.
// Close will close and remove an unreleased file and does
// nothing if it has been released.
// If it has been releases the new realm is responsible.
// to close and remove it.
type TempFile = file.TempFile

func NewTempFile(dir string, pattern string, fss ...vfs.FileSystem) (*TempFile, error) {
	return file.NewTempFile(dir, pattern, fss...)
}
