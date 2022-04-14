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
	"io"

	"github.com/open-component-model/ocm/pkg/errors"
)

type closableReader struct {
	reader io.Reader
}

func ReadCloser(r io.Reader) io.ReadCloser { return closableReader{r} }

func (r closableReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
func (r closableReader) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// NopWriteCloser returns a ReadCloser with a no-op Close method wrapping
// the provided Reader r.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

////////////////////////////////////////////////////////////////////////////////

type additionalCloser struct {
	reader           io.ReadCloser
	additionalCloser io.Closer
}

var _ io.ReadCloser = (*additionalCloser)(nil)

func AddCloser(reader io.ReadCloser, closer io.Closer) io.ReadCloser {
	return &additionalCloser{
		reader:           reader,
		additionalCloser: closer,
	}
}

func (c *additionalCloser) Close() error {
	list := errors.ErrListf("synthesized artefact blob")
	list.Add(c.reader.Close())
	list.Add(c.additionalCloser.Close())
	return list.Result()
}

func (c *additionalCloser) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

////////////////////////////////////////////////////////////////////////////////

func BlobData(blob DataGetter, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return blob.Get()
}

func BlobReader(blob DataReader, err error) (io.ReadCloser, error) {
	if err != nil {
		return nil, err
	}
	return blob.Reader()
}
