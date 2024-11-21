package accessio

import (
	"bytes"
	"io"
	"sync"

	"github.com/mandelsoft/goutils/general"
)

type ChunkedReader struct {
	lock   sync.Mutex
	reader io.Reader
	buffer *bytes.Buffer
	size   uint64
	chunk  uint64
	read   uint64
	err    error

	preread uint
}

var _ io.Reader = (*ChunkedReader)(nil)

func NewChunkedReader(r io.Reader, chunk uint64, preread ...uint) *ChunkedReader {
	return &ChunkedReader{
		reader:  r,
		size:    chunk,
		preread: general.OptionalDefaulted(8096, preread...),
	}
}

func (c *ChunkedReader) Read(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.read == c.size {
		return 0, io.EOF
	}
	if c.read+uint64(len(p)) > c.size {
		p = p[:c.size-c.read] // read at most rest of chunk size
	}
	if c.buffer != nil && c.buffer.Len() > 0 {
		// first, consume from buffer
		n, _ := c.buffer.Read(p)
		c.read += uint64(n)
		if c.buffer.Len() == 0 {
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
	c.read += uint64(n)
	return c.report(n, err)
}

func (c *ChunkedReader) report(n int, err error) (int, error) {
	if err == nil && c.read >= c.size {
		err = io.EOF
	}
	return n, err
}

func (c *ChunkedReader) ChunkDone() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.read >= c.size
}

func (c *ChunkedReader) Next() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.read < c.size || c.err != nil {
		return false
	}

	if c.buffer == nil {
		// cannot assume that read with size 0 returns EOF as proposed
		// by io.Reader.Read (see bytes.Buffer.Read).
		// Therefore, we really have to read something.

		var buf = make([]byte, c.preread, c.preread)
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
	c.chunk++
	return true
}
