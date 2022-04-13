// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package accessio

import (
	"bytes"
	"io"
	"sync"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
)

var ErrClosed = errors.ErrClosed()
var ErrReadOnly = errors.ErrReadOnly()

const KIND_BLOB = "blob"
const KIND_MEDIATYPE = "media type"

func ErrBlobNotFound(digest digest.Digest) error {
	return errors.ErrNotFound(KIND_BLOB, digest.String())
}

func IsErrBlobNotFound(err error) bool {
	return errors.IsErrNotFoundKind(err, KIND_BLOB)
}

type DataGetter interface {
	// Get returns the content as byte array
	Get() ([]byte, error)
}

type DataReader interface {
	// Reader returns a reader to incrementally access byte stream content
	Reader() (io.ReadCloser, error)
}

////////////////////////////////////////////////////////////////////////////////

//  DataAccess describes the access to sequence of bytes
type DataAccess interface {
	DataGetter
	DataReader
}

type MimeType interface {
	// MimeType returns the mime type of the blob
	MimeType() string
}

//  BlobAccess describes the access to a blob
type BlobAccess interface {
	DataAccess
	DigestSource
	MimeType

	// DigestKnown reports whether digest is already known
	DigestKnown() bool
	// Size returns the blob size
	Size() int64
}

// TemporaryBlobAccess describes a blob with temporary allocated external resources.
// The will be releases, when the close method is called
type TemporaryBlobAccess interface {
	BlobAccess
	Close() error
}

type DigestSource interface {
	// Digest returns the blob digest
	Digest() digest.Digest
}

////////////////////////////////////////////////////////////////////////////////

type Access = interface{}

func CloseAccess(a Access) error {
	if c, ok := a.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type dataAccess struct {
	fs   vfs.FileSystem
	path string
}

func DataAccessForFile(fs vfs.FileSystem, path string) DataAccess {
	return &dataAccess{fs, path}
}

func (a *dataAccess) Get() ([]byte, error) {
	return vfs.ReadFile(a.fs, a.path)
}

func (a *dataAccess) Reader() (io.ReadCloser, error) {
	return a.fs.Open(a.path)
}

////////////////////////////////////////////////////////////////////////////////

type bytesAccess struct {
	data []byte
}

func DataAccessForBytes(data []byte) DataAccess {
	return &bytesAccess{data}
}

func (a *bytesAccess) Get() ([]byte, error) {
	return a.data, nil
}

func (a *bytesAccess) Reader() (io.ReadCloser, error) {
	return ReadCloser(bytes.NewReader(a.data)), nil
}

////////////////////////////////////////////////////////////////////////////////

type blobAccess struct {
	lock     sync.RWMutex
	digest   digest.Digest
	size     int64
	mimeType string
	access   DataAccess
}

const BLOB_UNKNOWN_SIZE = int64(-1)
const BLOB_UNKNOWN_DIGEST = digest.Digest("")

func BlobAccessForDataAccess(digest digest.Digest, size int64, mimeType string, access DataAccess) BlobAccess {
	return &blobAccess{
		digest:   digest,
		size:     size,
		mimeType: mimeType,
		access:   access,
	}
}

func BlobAccessForString(mimeType string, data string) BlobAccess {
	return BlobAccessForData(mimeType, []byte(data))
}

func BlobAccessForData(mimeType string, data []byte) BlobAccess {
	return &blobAccess{
		digest:   digest.FromBytes(data),
		size:     int64(len(data)),
		mimeType: mimeType,
		access:   DataAccessForBytes(data),
	}
}

func (b *blobAccess) Get() ([]byte, error) {
	return b.access.Get()
}

func (b *blobAccess) Reader() (io.ReadCloser, error) {
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
	if err == nil {
		defer reader.Close()
		count := NewCountingReader(reader)
		digest, err := digest.Canonical.FromReader(count)
		if err == nil {
			b.size = count.Size()
			b.digest = digest
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type mimeBlob struct {
	BlobAccess
	mimetype string
}

func BlobWithMimeType(mimeType string, blob BlobAccess) BlobAccess {
	return &mimeBlob{blob, mimeType}
}

func (b *mimeBlob) MimeType() string {
	return b.mimetype
}

////////////////////////////////////////////////////////////////////////////////

type fileBlobAccess struct {
	dataAccess
	mimeType string
}

var _ BlobAccess = (*fileBlobAccess)(nil)

func BlobAccessForFile(mimeType string, path string, fs vfs.FileSystem) BlobAccess {
	return &fileBlobAccess{
		mimeType:   mimeType,
		dataAccess: dataAccess{fs, path},
	}
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

////////////////////////////////////////////////////////////////////////////////

type blobNopCloser struct {
	BlobAccess
}

func BlobNopCloser(blob BlobAccess) TemporaryBlobAccess {
	return &blobNopCloser{blob}
}

func (b *blobNopCloser) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type TemporaryFileSystemBlobAccess interface {
	TemporaryBlobAccess
	FileSystem() vfs.FileSystem
	Path() string
}

type temporaryBlob struct {
	BlobAccess
	temp       vfs.File
	filesystem vfs.FileSystem
}

func TempFileBlobAccess(mime string, fs vfs.FileSystem, temp vfs.File) TemporaryFileSystemBlobAccess {
	return &temporaryBlob{
		BlobAccess: BlobAccessForFile(mime, temp.Name(), fs),
		filesystem: fs,
		temp:       temp,
	}
}

func (a *temporaryBlob) Close() error {
	if a.temp != nil {
		list := errors.ErrListf("temporary blob")
		list.Add(a.temp.Close())
		list.Add(a.filesystem.Remove(a.temp.Name()))
		a.temp = nil
		return list.Result()
	}
	return nil
}

func (a *temporaryBlob) FileSystem() vfs.FileSystem {
	return a.filesystem
}

func (a *temporaryBlob) Path() string {
	return a.temp.Name()
}

// TempFile holds a temporary file that should be kept open.
// Close should neven be called directly.
// It can be passed to another responsibility realm be Release
// are transformed into a TemporaryBlobAccess.
// Close will close and remove an unreleased file and does
// nothing if it has been released.
// If it has been releases the new realm is responsible.
// to close and remove it.
type TempFile struct {
	lock       sync.Mutex
	temp       vfs.File
	filesystem vfs.FileSystem
}

func NewTempFile(fs vfs.FileSystem, dir string, pattern string) (*TempFile, error) {
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

func (t *TempFile) AsBlob(mime string) TemporaryFileSystemBlobAccess {
	return TempFileBlobAccess(mime, t.filesystem, t.Release())
}

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
