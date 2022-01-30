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

//  DataAccess describes the access to sequence of bytes
type DataAccess interface {
	// Get returns the content of the blob as byte array
	Get() ([]byte, error)
	// Reader returns a reader to incrementally access the blob content
	Reader() (io.ReadCloser, error)
}

//  BlobAccess describes the access to a blob
type BlobAccess interface {
	DataAccess

	// MimeType returns the mime type of the blob
	MimeType() string
	// Digest returns the blob digest
	Digest() digest.Digest
	// Size returns the blob size
	Size() int64
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

const BLOB_UNKNOWN_SIZE = -1
const BLOB_UNKNOWN_DIGEST = digest.Digest("")

func BlobAccessForFile(digest digest.Digest, size int64, mimeType string, access DataAccess) BlobAccess {
	return &blobAccess{
		digest:   digest,
		size:     size,
		mimeType: mimeType,
		access:   access,
	}
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
