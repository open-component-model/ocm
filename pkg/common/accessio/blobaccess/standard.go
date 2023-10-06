package blobaccess

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess/spi"
	"github.com/open-component-model/ocm/pkg/common/accessio/refmgmt"
	"github.com/open-component-model/ocm/pkg/common/iotools"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/opencontainers/go-digest"
)

var ErrClosed = refmgmt.ErrClosed

func ErrBlobNotFound(digest digest.Digest) error {
	return errors.ErrNotFound(KIND_BLOB, digest.String())
}

func IsErrBlobNotFound(err error) bool {
	return errors.IsErrNotFoundKind(err, KIND_BLOB)
}

////////////////////////////////////////////////////////////////////////////////

// Validatable is an optional interface for DataAccess
// implementations or any other object, which might reach
// an error state. The error can then be queried with
// the method ErrorProvider.Validate.
// This is used to support objects with access methods not
// returning an error. If the object is not valid,
// those methods return an unknown/default state, but
// the object should be queryable for its state.
type Validatable = utils.Validatable

// Validate checks whether a blob access
// is in error state. If yes, an appropriate
// error is returned.
func Validate(o BlobAccess) error {
	return utils.ValidateObject(o)
}

////////////////////////////////////////////////////////////////////////////////

type blobAccess struct {
	lock     sync.RWMutex
	digest   digest.Digest
	size     int64
	mimeType string
	closed   atomic.Bool
	access   DataAccess
}

func (b *blobAccess) Close() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if !b.closed.Load() {
		tmp := b.access
		b.closed.Store(true)
		return tmp.Close()
	}
	return ErrClosed
}

func (b *blobAccess) Get() ([]byte, error) {
	if b.closed.Load() {
		return nil, ErrClosed
	}
	return b.access.Get()
}

func (b *blobAccess) Reader() (io.ReadCloser, error) {
	if b.closed.Load() {
		return nil, ErrClosed
	}
	return b.access.Reader()
}

func (b *blobAccess) MimeType() string {
	return b.mimeType
}

func (b *blobAccess) DigestKnown() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.digest != ""
}

func (b *blobAccess) Digest() digest.Digest {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.digest == "" {
		b.update()
	}
	return b.digest
}

func (b *blobAccess) Size() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.size < 0 {
		b.update()
	}
	return b.size
}

func (b *blobAccess) update() error {
	reader, err := b.Reader()
	if err != nil {
		return err
	}

	defer reader.Close()
	count := iotools.NewCountingReader(reader)

	digest, err := digest.Canonical.FromReader(count)
	if err != nil {
		return err
	}

	b.size = count.Size()
	b.digest = digest

	return nil
}

func ForString(mime string, data string) BlobAccess {
	return ForData(mime, []byte(data))
}

func ForData(mime string, data []byte) BlobAccess {
	return spi.NewBlobAccessForBase(&blobAccess{
		digest:   digest.FromBytes(data),
		size:     int64(len(data)),
		mimeType: mime,
		access:   DataAccessForBytes(data),
	})
}

type _blobAccess = BlobAccess

////////////////////////////////////////////////////////////////////////////////

type fileBlobAccess struct {
	dataAccess
	mimeType string
}

var (
	_ BlobAccess   = (*fileBlobAccess)(nil)
	_ FileLocation = (*fileBlobAccess)(nil)
)

func (f *fileBlobAccess) FileSystem() vfs.FileSystem {
	return f.fs
}

func (f *fileBlobAccess) Path() string {
	return f.path
}

func (f *fileBlobAccess) Dup() (BlobAccess, error) {
	return f, nil
}

func (f *fileBlobAccess) Size() int64 {
	size := BLOB_UNKNOWN_SIZE
	fi, err := f.fs.Stat(f.path)
	if err == nil {
		size = fi.Size()
	}
	return size
}

func (f *fileBlobAccess) MimeType() string {
	return f.mimeType
}

func (f *fileBlobAccess) DigestKnown() bool {
	return false
}

func (f *fileBlobAccess) Digest() digest.Digest {
	r, err := f.Reader()
	if err != nil {
		return ""
	}
	defer r.Close()
	d, err := digest.FromReader(r)
	if err != nil {
		return ""
	}
	return d
}

func ForFile(mime string, path string, fss ...vfs.FileSystem) BlobAccess {
	return spi.NewBlobAccessForBase(&fileBlobAccess{
		mimeType:   mime,
		dataAccess: dataAccess{fs: utils.FileSystem(fss...), path: path},
	})
}

func ForFileWithCloser(closer io.Closer, mime string, path string, fss ...vfs.FileSystem) BlobAccess {
	return spi.NewBlobAccessForBase(&fileBlobAccess{
		mimeType:   mime,
		dataAccess: dataAccess{fs: utils.FileSystem(fss...), path: path},
	}, closer)
}

////////////////////////////////////////////////////////////////////////////////

// AnnotatedBlobAccess provides access to the original underlying data source.
type AnnotatedBlobAccess[T DataAccess] interface {
	_blobAccess
	Source() T
}

type annotatedBlobAccessView[T DataAccess] struct {
	_blobAccess
	access T
}

func (a *annotatedBlobAccessView[T]) Dup() (BlobAccess, error) {
	b, err := a._blobAccess.Dup()
	if err != nil {
		return nil, err
	}
	return &annotatedBlobAccessView[T]{
		_blobAccess: b,
		access:      a.access,
	}, nil
}

func (a *annotatedBlobAccessView[T]) Source() T {
	return a.access
}

// ForDataAccess wraps the general access object into a blob access.
// It closes the wrapped access, if closed.
func ForDataAccess[T DataAccess](digest digest.Digest, size int64, mimeType string, access T) AnnotatedBlobAccess[T] {
	a := &blobAccess{
		digest:   digest,
		size:     size,
		mimeType: mimeType,
		access:   access,
	}

	return &annotatedBlobAccessView[T]{
		_blobAccess: spi.NewBlobAccessForBase(a),
		access:      access,
	}
}

////////////////////////////////////////////////////////////////////////////////

type temporaryFileBlob struct {
	_blobAccess
	lock       sync.Mutex
	path       string
	file       vfs.File
	filesystem vfs.FileSystem
}

var (
	_ BlobAccessBase = (*temporaryFileBlob)(nil)
	_ FileLocation   = (*temporaryFileBlob)(nil)
)

func (b *temporaryFileBlob) Validate() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.path == "" {
		return ErrClosed
	}
	ok, err := vfs.Exists(b.filesystem, b.path)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNotFound("file", b.path)
	}
	return nil
}

func (b *temporaryFileBlob) Close() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.path != "" {
		list := errors.ErrListf("temporary blob")
		if b.file != nil {
			list.Add(b.file.Close())
		}
		list.Add(b.filesystem.Remove(b.path))
		b.path = ""
		b.file = nil
		b._blobAccess = nil
		return list.Result()
	}
	return nil
}

func (b *temporaryFileBlob) FileSystem() vfs.FileSystem {
	return b.filesystem
}

func (b *temporaryFileBlob) Path() string {
	return b.path
}

func ForTemporaryFile(mime string, temp vfs.File, fss ...vfs.FileSystem) BlobAccess {
	return spi.NewBlobAccessForBase(&temporaryFileBlob{
		_blobAccess: ForFile(mime, temp.Name(), fss...),
		filesystem:  utils.FileSystem(fss...),
		path:        temp.Name(),
		file:        temp,
	})
}

func ForTemporaryFilePath(mime string, temp string, fss ...vfs.FileSystem) BlobAccess {
	return spi.NewBlobAccessForBase(&temporaryFileBlob{
		_blobAccess: ForFile(mime, temp, fss...),
		filesystem:  utils.FileSystem(fss...),
		path:        temp,
	})
}

////////////////////////////////////////////////////////////////////////////////

type mimeBlob struct {
	_blobAccess
	mimetype string
}

// BlobWithMimeType changes the mime type for a blob access
// by wrapping the given blob access. It does NOT provide
// a new view for the given blob access, so closing the resulting
// blob access will directly close the backing blob access.
func BlobWithMimeType(mimeType string, blob BlobAccess) BlobAccess {
	return &mimeBlob{blob, mimeType}
}

func (b *mimeBlob) Dup() (BlobAccess, error) {
	n, err := b._blobAccess.Dup()
	if err != nil {
		return nil, err
	}
	return &mimeBlob{n, b.mimetype}, nil
}

func (b *mimeBlob) MimeType() string {
	return b.mimetype
}

////////////////////////////////////////////////////////////////////////////////

type blobNopCloser struct {
	_blobAccess
}

func NonClosable(blob BlobAccess) BlobAccess {
	return &blobNopCloser{blob}
}

func (b *blobNopCloser) Close() error {
	return nil
}
