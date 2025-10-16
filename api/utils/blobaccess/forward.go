package blobaccess

import (
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/dirtree"
	"ocm.software/ocm/api/utils/blobaccess/dockerdaemon"
	"ocm.software/ocm/api/utils/blobaccess/dockermulti"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/blobaccess/helm"
	"ocm.software/ocm/api/utils/blobaccess/maven"
	"ocm.software/ocm/api/utils/blobaccess/ociartifact"
	"ocm.software/ocm/api/utils/blobaccess/wget"
)

///////////
// Standard
///////////

// DataAccessForData wraps a bytes slice into a DataAccess.
func DataAccessForData(data []byte, origin ...string) DataSource {
	return blobaccess.DataAccessForData(data, origin...)
}

func DataAccessForString(data string, origin ...string) DataSource {
	return blobaccess.DataAccessForString(data, origin...)
}

// ForString wraps a string into a BlobAccess, which does not need a close.
func ForString(mime string, data string) BlobAccess {
	return blobaccess.ForString(mime, data)
}

func ProviderForString(mime, data string) BlobAccessProvider {
	return blobaccess.ProviderForString(mime, data)
}

// ForData wraps data into a BlobAccess, which does not need a close.
func ForData(mime string, data []byte) BlobAccess {
	return blobaccess.ForData(mime, data)
}

func ProviderForData(mime string, data []byte) BlobAccessProvider {
	return blobaccess.ProviderForData(mime, data)
}

///////////
// File
///////////

func DataAccessForFile(fs vfs.FileSystem, path string) DataAccess {
	return file.DataAccess(fs, path)
}

func ForFile(mime string, path string, fss ...vfs.FileSystem) BlobAccess {
	return file.BlobAccess(mime, path, fss...)
}

func ProviderForFile(mime string, path string, fss ...vfs.FileSystem) BlobAccessProvider {
	return file.Provider(mime, path, fss...)
}

func ForFileWithCloser(closer io.Closer, mime string, path string, fss ...vfs.FileSystem) BlobAccess {
	return file.BlobAccessWithCloser(closer, mime, path, fss...)
}

func ForTemporaryFile(mime string, temp vfs.File, opts ...file.Option) BlobAccess {
	return file.BlobAccessForTemporaryFile(mime, temp, opts...)
}

func ForTemporaryFilePath(mime string, temp string, opts ...file.Option) BlobAccess {
	return file.BlobAccessForTemporaryFilePath(mime, temp, opts...)
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

///////////
// DirTree
///////////

func DataAccessForDirTree(path string, opts ...dirtree.Option) (DataAccess, error) {
	return dirtree.DataAccess(path, opts...)
}

func ForDirTree(path string, opts ...dirtree.Option) (BlobAccess, error) {
	return dirtree.BlobAccess(path, opts...)
}

func ProviderForDirTree(path string, opts ...dirtree.Option) BlobAccessProvider {
	return dirtree.Provider(path, opts...)
}

///////////
// Docker Daemon
///////////

func ForImageFromDockerDaemon(name string, opts ...dockerdaemon.Option) (BlobAccess, string, error) {
	return dockerdaemon.BlobAccess(name, opts...)
}

func ProviderForImageFromDockerDaemon(name string, opts ...dockerdaemon.Option) BlobAccessProvider {
	return dockerdaemon.Provider(name, opts...)
}

///////////
// Docker Multi
///////////

func ForMultiImageFromDockerDaemon(opts ...dockermulti.Option) (BlobAccess, error) {
	return dockermulti.BlobAccess(opts...)
}

func ProviderForMultiImageFromDockerDaemon(opts ...dockermulti.Option) BlobAccessProvider {
	return dockermulti.Provider(opts...)
}

///////////
// Helm Chart
///////////

func ForHelmChart(path string, opts ...helm.Option) (blob BlobAccess, name, version string, err error) {
	return helm.BlobAccess(path, opts...)
}

func ProviderForHelmChart(path string, opts ...helm.Option) BlobAccessProvider {
	return helm.Provider(path, opts...)
}

///////////
// Maven
///////////

func DataAccessForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...maven.Option) (DataAccess, error) {
	return maven.DataAccess(repo, groupId, artifactId, version, opts...)
}

func ForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...maven.Option) (BlobAccess, error) {
	return maven.BlobAccess(repo, groupId, artifactId, version, opts...)
}

func ForMavenCoords(repo *maven.Repository, coords *maven.Coordinates, opts ...maven.Option) (BlobAccess, error) {
	return maven.BlobAccessForCoords(repo, coords, opts...)
}

func ProviderForMaven(repo *maven.Repository, groupId, artifactId, version string, opts ...maven.Option) BlobAccessProvider {
	return maven.Provider(repo, groupId, artifactId, version, opts...)
}

///////////
// OCI Artifact
///////////

func ForOCIArtifact(refname string, opts ...ociartifact.Option) (BlobAccess, string, error) {
	return ociartifact.BlobAccess(refname, opts...)
}

func ProviderForOCIArtifact(name string, opts ...ociartifact.Option) BlobAccessProvider {
	return ociartifact.Provider(name, opts...)
}

///////////
// WGet
///////////

func DataAccessForWget(url string, opts ...wget.Option) (DataAccess, error) {
	return wget.DataAccess(url, opts...)
}

func ForWget(url string, opts ...wget.Option) (_ BlobAccess, rerr error) {
	return wget.BlobAccess(url, opts...)
}

func ProviderForWget(url string, opts ...wget.Option) BlobAccessProvider {
	return wget.Provider(url, opts...)
}

////////////////////////////////////////////////////////////////////////////////

type _blobAccess = BlobAccess

// AnnotatedBlobAccess provides access to the original underlying data source.
type AnnotatedBlobAccess[T bpi.DataAccess] interface {
	_blobAccess
	Source() T
}

func ForDataAccess[T bpi.DataAccess](digest digest.Digest, size int64, mimeType string, access T) AnnotatedBlobAccess[T] {
	return blobaccess.ForDataAccess(digest, size, mimeType, access)
}

func ProviderForBlobAccess(blob bpi.BlobAccess) BlobAccessProvider {
	return blobaccess.ProviderForBlobAccess(blob)
}

////////////////////////////////////////////////////////////////////////////////

func BlobData(blob DataGetter, err ...error) ([]byte, error) {
	return blobaccess.BlobData(blob, err...)
}

func BlobReader(blob DataReader, err ...error) (io.ReadCloser, error) {
	return blobaccess.BlobReader(blob, err...)
}

func Digest(access DataAccess) (digest.Digest, error) {
	return blobaccess.Digest(access)
}

func WithCompression(blob BlobAccess) (BlobAccess, error) {
	return blobaccess.WithCompression(blob)
}

func WithDecompression(blob BlobAccess) (BlobAccess, error) {
	return blobaccess.WithDecompression(blob)
}

func DataAccessForReaderFunction(reader func() (io.ReadCloser, error), origin string) DataAccess {
	return blobaccess.DataAccessForReaderFunction(reader, origin)
}

////////////////////////////////////////////////////////////////////////////////

type GenericData = blobaccess.GenericData

type GenericDataGetter = blobaccess.GenericDataGetter

const KIND_DATASOURCE = blobaccess.KIND_DATASOURCE

// GetData provides data as byte sequence from some generic
// data sources like byte arrays, strings, DataReader and
// DataGetters. This means we can pass all BlobAccess or DataAccess
// objects.
// If no an unknown data source is passes an ErrInvalid(KIND_DATASOURCE)
// is returned.
func GetData(src GenericData) ([]byte, error) {
	return blobaccess.GetData(src)
}

// GetGenericData evaluates some input provided by well-known
// types or interfaces and provides some data output
// by mapping the input to either a byte sequence or
// some specialized object.
// If the input type is not known an ErrInvalid(KIND_DATASOURCE)
// // is returned.
// In extension to GetData, it additionally evaluates the interface
// GenericDataGetter to map the input to some evaluated object.
func GetGenericData(src GenericData) (interface{}, error) {
	return blobaccess.GetGenericData(src)
}
