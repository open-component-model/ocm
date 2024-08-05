package file

import (
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/iotools"
)

type (
	_nopCloser  = iotools.NopCloser
	_blobAccess = bpi.BlobAccess
)

////////////////////////////////////////////////////////////////////////////////

type fileDataAccess struct {
	_nopCloser
	fs   vfs.FileSystem
	path string
}

var (
	_ bpi.DataSource  = (*fileDataAccess)(nil)
	_ bpi.Validatable = (*fileDataAccess)(nil)
)

func DataAccess(fs vfs.FileSystem, path string) bpi.DataAccess {
	return &fileDataAccess{fs: fs, path: path}
}

func (a *fileDataAccess) Get() ([]byte, error) {
	data, err := vfs.ReadFile(a.fs, a.path)
	if err != nil {
		return nil, errors.Wrapf(err, "file %q", a.path)
	}
	return data, nil
}

func (a *fileDataAccess) Reader() (io.ReadCloser, error) {
	file, err := a.fs.Open(a.path)
	if err != nil {
		return nil, errors.Wrapf(err, "file %q", a.path)
	}
	return file, nil
}

// Validate checks if the access is valid, meaning
// it can provide data. Here, this means
// that the file exists.
func (a *fileDataAccess) Validate() error {
	ok, err := vfs.Exists(a.fs, a.path)
	if err != nil {
		return err
	}
	if !ok {
		return errors.ErrNotFound("file", a.path)
	}
	return nil
}

func (a *fileDataAccess) Origin() string {
	return a.path
}

////////////////////////////////////////////////////////////////////////////////

type fileBlobAccess struct {
	fileDataAccess
	mimeType string
}

var (
	_ bpi.BlobAccess   = (*fileBlobAccess)(nil)
	_ bpi.FileLocation = (*fileBlobAccess)(nil)
)

func (f *fileBlobAccess) FileSystem() vfs.FileSystem {
	return f.fs
}

func (f *fileBlobAccess) Path() string {
	return f.path
}

func (f *fileBlobAccess) Dup() (bpi.BlobAccess, error) {
	return f, nil
}

func (f *fileBlobAccess) Size() int64 {
	size := bpi.BLOB_UNKNOWN_SIZE
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

// BlobAccess wraps a file path into a BlobAccess, which does not need a close.
func BlobAccess(mime string, path string, fss ...vfs.FileSystem) bpi.BlobAccess {
	return &fileBlobAccess{
		mimeType:       mime,
		fileDataAccess: fileDataAccess{fs: utils.FileSystem(fss...), path: path},
	}
}

func Provider(mime string, path string, fss ...vfs.FileSystem) bpi.BlobAccessProvider {
	return bpi.BlobAccessProviderFunction(func() (bpi.BlobAccess, error) {
		return BlobAccess(mime, path, fss...), nil
	})
}

type fileBlobAccessView struct {
	_blobAccess
	access *fileDataAccess
}

var (
	_ bpi.BlobAccess   = (*fileBlobAccessView)(nil)
	_ bpi.FileLocation = (*fileBlobAccessView)(nil)
)

func (f *fileBlobAccessView) Dup() (bpi.BlobAccess, error) {
	b, err := f._blobAccess.Dup()
	if err != nil {
		return nil, err
	}
	return &fileBlobAccessView{b, f.access}, nil
}

func (f *fileBlobAccessView) FileSystem() vfs.FileSystem {
	return f.access.fs
}

func (f *fileBlobAccessView) Path() string {
	return f.access.path
}

func BlobAccessWithCloser(closer io.Closer, mime string, path string, fss ...vfs.FileSystem) bpi.BlobAccess {
	fb := &fileBlobAccess{fileDataAccess{fs: utils.FileSystem(fss...), path: path}, mime}
	return &fileBlobAccessView{
		bpi.NewBlobAccessForBase(fb, closer),
		&fb.fileDataAccess,
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
	_ bpi.BlobAccessBase = (*temporaryFileBlob)(nil)
	_ bpi.FileLocation   = (*temporaryFileBlob)(nil)
)

// Validate checks if the access is valid, meaning
// it can provide data. Here, this means
// that the file exists.
func (b *temporaryFileBlob) Validate() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.path == "" {
		return bpi.ErrClosed
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

func BlobAccessForTemporaryFile(mime string, temp vfs.File, opts ...Option) bpi.BlobAccess {
	eff := optionutils.EvalOptions(opts...)
	t := &temporaryFileBlob{
		_blobAccess: BlobAccess(mime, temp.Name(), eff.FileSystem),
		filesystem:  utils.FileSystem(eff.FileSystem),
		path:        temp.Name(),
		file:        temp,
	}
	// TODO: handle FileLocation interface in combination with partially set meta data.
	if eff.Digest != "" || eff.GetSize() != bpi.BLOB_UNKNOWN_SIZE {
		return bpi.NewBlobAccessForBase(bpi.BaseAccessForDataAccessAndMeta(mime, t, eff.Digest, eff.GetSize()))
	}
	return bpi.NewBlobAccessForBase(t)
}

func BlobAccessForTemporaryFilePath(mime string, temp string, opts ...Option) bpi.BlobAccess {
	eff := optionutils.EvalOptions(opts...)
	return bpi.NewBlobAccessForBase(bpi.BaseAccessForDataAccessAndMeta(mime, &temporaryFileBlob{
		_blobAccess: BlobAccess(mime, temp, eff.FileSystem),
		filesystem:  utils.FileSystem(eff.FileSystem),
		path:        temp,
	}, eff.Digest, eff.GetSize()))
}

////////////////////////////////////////////////////////////////////////////////

// TempFile holds a temporary file that should be kept open.
// Close should never be called directly.
// It can be passed to another responsibility realm by calling Release-
// For example to be transformed into a TemporaryBlobAccess.
// Close will close and remove an unreleased file and does
// nothing if it has been released.
// If it has been released the new realm is responsible.
// to close and remove it.
type TempFile struct {
	lock       sync.Mutex
	temp       vfs.File
	filesystem vfs.FileSystem
}

func NewTempFile(dir string, pattern string, fss ...vfs.FileSystem) (*TempFile, error) {
	fs := utils.FileSystem(fss...)
	temp, err := vfs.TempFile(fs, dir, pattern)
	if err != nil {
		return nil, err
	}
	return &TempFile{
		temp:       temp,
		filesystem: fs,
	}, nil
}

func (t *TempFile) Name() string {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.temp.Name()
}

func (t *TempFile) FileSystem() vfs.FileSystem {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.filesystem
}

// Release passes the responsibility for closing and removing
// the temporary file to another realm. After calling this method
// the TempFile object will not handle these operations anymore, if it is closed.
func (t *TempFile) Release() vfs.File {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.temp != nil {
		t.temp.Sync()
	}
	tmp := t.temp
	t.temp = nil
	return tmp
}

func (t *TempFile) Writer() io.Writer {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.temp
}

func (t *TempFile) Sync() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.temp.Sync()
}

func (t *TempFile) AsBlob(mime string) bpi.BlobAccess {
	return BlobAccessForTemporaryFile(mime, t.Release(), WithFileSystem(t.filesystem))
}

// Close closes and removes the temporary file as long it has not
// been released before by calling Release.
func (t *TempFile) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.temp != nil {
		name := t.temp.Name()
		t.temp.Close()
		t.temp = nil
		return t.filesystem.Remove(name)
	}
	return nil
}
