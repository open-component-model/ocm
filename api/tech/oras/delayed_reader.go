package oras

import (
	"io"
	"sync"
)

// delayedReader sets up a reader that only fetches a blob
// upon explicit reading request, otherwise, it stores the
// way of getting the reader.
type delayedReader struct {
	open   func() (io.ReadCloser, error)
	rc     io.ReadCloser
	closed bool
	rw     sync.Mutex
}

func newDelayedReader(open func() (io.ReadCloser, error)) (*delayedReader, error) {
	return &delayedReader{
		open: open,
	}, nil
}

func (d *delayedReader) Read(p []byte) (n int, err error) {
	d.rw.Lock()
	defer d.rw.Unlock()

	if d.closed {
		return 0, io.EOF
	}

	reader, err := d.reader()
	if err != nil {
		return 0, err
	}

	return reader.Read(p)
}

func (d *delayedReader) reader() (io.ReadCloser, error) {
	d.rw.Lock()
	defer d.rw.Unlock()

	if d.rc != nil {
		return d.rc, nil
	}

	rc, err := d.open()
	if err != nil {
		return nil, err
	}

	d.rc = rc
	return rc, nil
}

func (d *delayedReader) Close() error {
	d.rw.Lock()
	defer d.rw.Unlock()

	if d.closed {
		return nil
	}

	// we close regardless of an error
	d.closed = true
	return d.rc.Close()
}
