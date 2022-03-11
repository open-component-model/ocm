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

package compression

import (
	"bytes"
	"io"
)

type MatchReader interface {
	io.Reader
	Reset()
}

type matchReader struct {
	read    []byte
	buffer  *bytes.Buffer
	reader  io.Reader
	current io.Reader
}

var _ MatchReader = (*matchReader)(nil)

func NewMatchReader(r io.Reader) *matchReader {
	return &matchReader{
		buffer:  bytes.NewBuffer(nil),
		reader:  r,
		current: r,
	}
}

func (r *matchReader) Read(buf []byte) (int, error) {
	n, err := r.current.Read(buf)
	if n > 0 {
		_, err = r.buffer.Write(buf[:n])
	}
	return n, err
}

func (r *matchReader) Reset() {
	if r.buffer.Len() > 0 {
		if r.buffer.Len() > len(r.read) {
			r.read = r.buffer.Bytes()
		}
		r.buffer = bytes.NewBuffer(nil)
		r.current = io.MultiReader(bytes.NewBuffer(r.read), r.reader)
	}
}

func (r *matchReader) Reader() io.Reader {
	r.Reset()
	r.reader = nil
	return r.current
}
