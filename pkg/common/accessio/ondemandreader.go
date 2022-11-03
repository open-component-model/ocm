// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package accessio

import (
	"io"
	"sync"
)

type ReaderProvider interface {
	Reader() (io.ReadCloser, error)
}

type OnDemandReader struct {
	lock     sync.Mutex
	provider ReaderProvider
	reader   io.ReadCloser
}

var _ io.Reader = (*OnDemandReader)(nil)

func NewOndemandReader(p ReaderProvider) io.ReadCloser {
	return &OnDemandReader{provider: p}
}

func (o *OnDemandReader) Read(p []byte) (n int, err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.reader == nil {
		r, err := o.provider.Reader()
		if err != nil {
			return 0, err
		}
		o.reader = r
	}
	return o.reader.Read(p)
}

func (o *OnDemandReader) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.reader == nil {
		return nil
	}
	return o.reader.Close()
}
