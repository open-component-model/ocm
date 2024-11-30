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
	lock    sync.Mutex
	reader  io.Reader
	buffer  *bytes.Buffer
	size    int64
	chunkNo int
	read    int64
	err     error

	preread uint
}

var _ io.Reader = (*ChunkedReader)(nil)

func NewChunkedReader(r io.Reader, chunkSize int64, preread ...uint) *ChunkedReader {
	return &ChunkedReader{
		reader:  r,
		size:    chunkSize,
		preread: general.OptionalDefaulted(8096, preread...),
	}
}

func (c *ChunkedReader) Read(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.read >= c.size {
		return 0, io.EOF
	}
	if c.read+int64(len(p)) > c.size {
		p = p[:c.size-c.read] // read at most rest of chunk size
	}
	if c.buffer != nil && c.buffer.Len() > 0 {
		// first, consume from buffer
		n, err := c.buffer.Read(p)
		c.read += int64(n)
		if err != nil { // the only error returned is io.EOF
			c.buffer = nil
		}
		return c.report(n, nil)
	} else {
		c.buffer = nil
	}

	if c.err != nil {
		return 0, c.err
	}
	n, err = c.reader.Read(p)
	c.read += int64(n)
	c.err = err
	return c.report(n, err)
}

func (c *ChunkedReader) report(n int, err error) (int, error) {
	if err == nil && c.read >= c.size {
		err = io.EOF
	}
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

	return c.read >= c.size || c.err != nil
}

func (c *ChunkedReader) Next() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.read < c.size || c.err != nil {
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

	c.read = 0
	c.chunkNo++
	return true
}
