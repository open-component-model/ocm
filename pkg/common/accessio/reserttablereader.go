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
	"os"

	"github.com/open-component-model/ocm/pkg/errors"
)

type ResettableReader struct {
	orig  io.ReadCloser
	buf   Buffer
	count int
}

func NewResettableReader(orig io.ReadCloser, size int64) (*ResettableReader, error) {
	var buf Buffer
	var err error

	if size < 0 || size > 8192 {
		buf, err = NewFileBuffer()
		if err != nil {
			return nil, err
		}
	} else {
		buf = &memoryBuffer{}
	}
	return &ResettableReader{
		orig: orig, buf: buf,
	}, nil
}

func (b *ResettableReader) Read(out []byte) (int, error) {
	n, err := b.orig.Read(out)
	if n > 0 {
		return b.buf.Write(out[:n])
	}
	return n, err
}

func (b *ResettableReader) Close() error {
	// fmt.Printf("close resend buffer\n")
	b.buf.Close()
	b.buf = nil
	return b.orig.Close()
}

func (b *ResettableReader) Reset() (io.ReadCloser, error) {
	b.count++
	if b.buf.Len() <= 0 {
		return &prefixReader{
			nil,
			b,
		}, nil
	}
	r, err := b.buf.Reader()
	if err != nil {
		return nil, err
	}
	return &prefixReader{
		r,
		b,
	}, nil
}

type prefixReader struct {
	prefix io.ReadCloser
	resend *ResettableReader
}

func (p *prefixReader) Read(out []byte) (int, error) {
	if p.prefix != nil {
		n, err := p.prefix.Read(out)
		if err == nil {
			return n, nil
		}
		p.prefix.Close()
		p.prefix = nil
	}
	n, err := p.resend.Read(out)
	// fmt.Printf("blob read %d: %s\n", n, err)
	return n, err
}

func (p *prefixReader) Close() error {
	// fmt.Printf("close prefix reader\n")
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Buffer interface {
	Write(out []byte) (int, error)
	Reader() (io.ReadCloser, error)
	Len() int
	Close() error
}

type memoryBuffer struct {
	bytes.Buffer
}

var _ Buffer = (*memoryBuffer)(nil)

func (m *memoryBuffer) Reader() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.Bytes())), nil
}

func (m *memoryBuffer) Close() error {
	return nil
}

type fileBuffer struct {
	path string
	file *os.File
}

var _ Buffer = (*fileBuffer)(nil)

func NewFileBuffer() (*fileBuffer, error) {
	file, err := os.CreateTemp("", "ociblob*")
	if err != nil {
		return nil, err
	}
	return &fileBuffer{
		path: file.Name(),
		file: file,
	}, nil
}

func (b *fileBuffer) Write(out []byte) (int, error) {
	return b.file.Write(out)
}

func (b *fileBuffer) Reader() (io.ReadCloser, error) {
	return os.Open(b.path)
}

func (b *fileBuffer) Len() int {
	fi, err := b.file.Stat()
	if err != nil {
		return -1
	}
	return int(fi.Size())
}

func (b *fileBuffer) Close() error {
	return errors.ErrListf("closing file buffer").Add(b.file.Close(), os.Remove(b.path)).Result()
}
