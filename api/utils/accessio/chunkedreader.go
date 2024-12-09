package accessio

import (
	"bytes"
	"io"
	"sync"

	"github.com/mandelsoft/goutils/general"
)

// ChunkedReader splits a reader into several
// logical readers with a limited content size.
// Once the reader reaches its limits it provides
// a io.EOF.
// It can be continued by Calling Next, which returns
// whether a follow-up is required or not.
type ChunkedReader struct {
	lock      sync.Mutex
	reader    io.Reader
	buffer    *bytes.Buffer
	chunkSize int64
	chunkNo   int

	limited io.Reader
	err     error

	preread uint
}

var _ io.Reader = (*ChunkedReader)(nil)

func NewChunkedReader(r io.Reader, chunkSize int64, preread ...uint) *ChunkedReader {
	return &ChunkedReader{
		reader:    r,
		chunkSize: chunkSize,
		limited:   io.LimitReader(r, chunkSize),
		preread:   min(uint(chunkSize-1), general.OptionalDefaulted(8096, preread...)),
	}
}

func (c *ChunkedReader) Read(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return 0, c.err
	}
	if c.buffer != nil && c.buffer.Len() > 0 {
		// first, consume from buffer
		n, err := c.buffer.Read(p)
		if err != nil { // the only error returned is io.EOF
			c.buffer = nil
		}
		if n > 0 {
			return n, nil
		}
	} else {
		c.buffer = nil
	}

	n, err = c.limited.Read(p)
	c.err = err
	return n, err

}

func (c *ChunkedReader) ChunkNo() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.chunkNo
}

func (c *ChunkedReader) ChunkDone() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.err == io.EOF
}

func (c *ChunkedReader) Next() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != io.EOF {
		return false
	}

	// cannot assume that read with size 0 returns EOF as proposed
	// by io.Reader.Read (see bytes.Buffer.Read).
	// Therefore, we really have to read something.
	if c.buffer == nil {
		buf := make([]byte, c.preread)
		n, err := c.reader.Read(buf)
		c.err = err
		if n > 0 {
			c.buffer = bytes.NewBuffer(buf[:n])
		} else {
			if err == io.EOF {
				return false
			}
		}
	}

	c.chunkNo++
	c.limited = io.LimitReader(c.reader, c.chunkSize-int64(c.preread))
	return true
}
