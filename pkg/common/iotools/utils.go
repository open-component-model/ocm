// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package iotools

import (
	"io"
)

type NopCloser struct{}

type _nopCloser = NopCloser

func (NopCloser) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

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
