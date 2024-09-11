package accessio

import (
	"crypto"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/iotools"
)

const (
	// Deprecated: use blobaccess.BLOB_UNKNOWN_SIZE.
	BLOB_UNKNOWN_SIZE = blobaccess.BLOB_UNKNOWN_SIZE
	// Deprecated: use blobaccess.BLOB_UNKNOWN_DIGEST.
	BLOB_UNKNOWN_DIGEST = blobaccess.BLOB_UNKNOWN_DIGEST
)

const (
	// Deprecated: use blobaccess.KIND_BLOB.
	KIND_BLOB = blobaccess.KIND_BLOB
	// Deprecated: use blobaccess.KIND_MEDIATYPE.
	KIND_MEDIATYPE = blobaccess.KIND_MEDIATYPE
)

// Deprecated: use blobaccess.ErrBlobNotFound.
func ErrBlobNotFound(digest digest.Digest) error {
	return errors.ErrNotFound(blobaccess.KIND_BLOB, digest.String())
}

// Deprecated: use blobaccess.IsErrBlobNotFound.
func IsErrBlobNotFound(err error) bool {
	return errors.IsErrNotFoundKind(err, blobaccess.KIND_BLOB)
}

////////////////////////////////////////////////////////////////////////////////

// DataSource describes some data plus its origin.
// Deprecated: use blobaccess.DataSource.
type DataSource = blobaccess.DataSource

////////////////////////////////////////////////////////////////////////////////

// DataAccess describes the access to sequence of bytes.
// Deprecated: use blobaccess.DataAccess.
type DataAccess = blobaccess.DataAccess

// BlobAccess describes the access to a blob.
// Deprecated: use blobaccess.BlobAccess.
type BlobAccess = blobaccess.BlobAccess

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use blobaccess.DataAccessForReaderFunction.
func DataAccessForReaderFunction(reader func() (io.ReadCloser, error), origin string) blobaccess.DataAccess {
	return blobaccess.DataAccessForReaderFunction(reader, origin)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use blobaccess.DataAccessForFile.
func DataAccessForFile(fs vfs.FileSystem, path string) blobaccess.DataAccess {
	return file.DataAccess(fs, path)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use blobaccess.DataAccessForBytes.
func DataAccessForBytes(data []byte, origin ...string) blobaccess.DataSource {
	return blobaccess.DataAccessForData(data, origin...)
}

// Deprecated: use blobaccess.DataAccessForString.
func DataAccessForString(data string, origin ...string) blobaccess.DataSource {
	return blobaccess.DataAccessForData([]byte(data), origin...)
}

////////////////////////////////////////////////////////////////////////////////

// BlobWithMimeType changes the mime type for a blob access
// by wrapping the given blob access. It does NOT provide
// a new view for the given blob access, so closing the resulting
// blob access will directly close the backing blob access.
// Deprecated: use blobaccess.WithMimeType.
func BlobWithMimeType(mimeType string, blob blobaccess.BlobAccess) blobaccess.BlobAccess {
	return blobaccess.WithMimeType(mimeType, blob)
}

////////////////////////////////////////////////////////////////////////////////

// AnnotatedBlobAccess provides access to the original underlying data source.
// Deprecated: use blobaccess.AnnotatedBlobAccess.
type AnnotatedBlobAccess[T blobaccess.DataAccess] interface {
	blobaccess.BlobAccess
	Source() T
}

// BlobAccessForDataAccess wraps the general access object into a blob access.
// It closes the wrapped access, if closed.
// Deprecated: use blobaccess.ForDataAccess.
func BlobAccessForDataAccess[T blobaccess.DataAccess](digest digest.Digest, size int64, mimeType string, access T) blobaccess.AnnotatedBlobAccess[T] {
	return blobaccess.ForDataAccess[T](digest, size, mimeType, access)
}

// Deprecated: use blobaccess.ForString.
func BlobAccessForString(mimeType string, data string) blobaccess.BlobAccess {
	return blobaccess.ForData(mimeType, []byte(data))
}

// Deprecated: use blobaccess.ForData.
func BlobAccessForData(mimeType string, data []byte) blobaccess.BlobAccess {
	return blobaccess.ForData(mimeType, data)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use blobaccess.ForFile.
func BlobAccessForFile(mimeType string, path string, fss ...vfs.FileSystem) blobaccess.BlobAccess {
	return file.BlobAccess(mimeType, path, fss...)
}

// Deprecated: use blobaccess.ForFileWithCloser.
func BlobAccessForFileWithCloser(closer io.Closer, mimeType string, path string, fss ...vfs.FileSystem) blobaccess.BlobAccess {
	return file.BlobAccessWithCloser(closer, mimeType, path, fss...)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use blobaccess.ForTemporaryFile.
func BlobAccessForTemporaryFile(mime string, temp vfs.File, fss ...vfs.FileSystem) blobaccess.BlobAccess {
	return file.BlobAccessForTemporaryFile(mime, temp, file.WithFileSystem(fss...))
}

// Deprecated: use blobaccess.ForTemporaryFilePath.
func BlobAccessForTemporaryFilePath(mime string, temp string, fss ...vfs.FileSystem) blobaccess.BlobAccess {
	return file.BlobAccessForTemporaryFilePath(mime, temp, file.WithFileSystem(fss...))
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use blobaccess.NewTempFile.
func NewTempFile(fs vfs.FileSystem, dir string, pattern string) (*file.TempFile, error) {
	return file.NewTempFile(dir, pattern, fs)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use iotools.DigestReader.
type DigestReader = iotools.DigestReader

// Deprecated: use iotools.NewDefaultDigestReader.
func NewDefaultDigestReader(r io.Reader) *iotools.DigestReader {
	return iotools.NewDigestReaderWith(digest.Canonical, r)
}

// Deprecated: use iotools.NewDigestReaderWith.
func NewDigestReaderWith(algorithm digest.Algorithm, r io.Reader) *iotools.DigestReader {
	return iotools.NewDigestReaderWith(algorithm, r)
}

// Deprecated: use iotools.NewDigestReaderWithHash.
func NewDigestReaderWithHash(hash crypto.Hash, r io.Reader) *iotools.DigestReader {
	return iotools.NewDigestReaderWithHash(hash, r)
}

// Deprecated: use iotools.VerifyingReader.
func VerifyingReader(r io.ReadCloser, digest digest.Digest) io.ReadCloser {
	return iotools.VerifyingReader(r, digest)
}

// Deprecated: use iotools.VerifyingReaderWithHash.
func VerifyingReaderWithHash(r io.ReadCloser, hash crypto.Hash, digest string) io.ReadCloser {
	return iotools.VerifyingReaderWithHash(r, hash, digest)
}

// Deprecated: use blobaccess.Digest.
func Digest(access blobaccess.DataAccess) (digest.Digest, error) {
	return blobaccess.Digest(access)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use iotools.DigestWriter.
type DigestWriter = iotools.DigestWriter

// Deprecated: use iotools.NewDefaultDigestWriter.
func NewDefaultDigestWriter(w io.WriteCloser) *iotools.DigestWriter {
	return iotools.NewDefaultDigestWriter(w)
}

// Deprecated: use iotools.NewDigestWriterWith.
func NewDigestWriterWith(algorithm digest.Algorithm, w io.WriteCloser) *iotools.DigestWriter {
	return iotools.NewDigestWriterWith(algorithm, w)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: use blobaccess.BlobData.
func BlobData(blob blobaccess.DataGetter, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return blob.Get()
}

// Deprecated: use blobaccess.BlobReader.
func BlobReader(blob blobaccess.DataReader, err error) (io.ReadCloser, error) {
	if err != nil {
		return nil, err
	}
	return blob.Reader()
}

// Deprecated: use utils.FileSystem.
func FileSystem(fss ...vfs.FileSystem) vfs.FileSystem {
	return utils.FileSystem(fss...)
}

// Deprecated: use utils.DefaultedFileSystem.
func DefaultedFileSystem(def vfs.FileSystem, fss ...vfs.FileSystem) vfs.FileSystem {
	return utils.DefaultedFileSystem(def, fss...)
}

// Deprecated: use iotools.AddReaderCloser.
func AddCloser(reader io.ReadCloser, closer io.Closer, msg ...string) io.ReadCloser {
	return iotools.AddReaderCloser(reader, closer, sliceutils.AsAny(msg)...)
}
