package accessio

import (
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"
)

// ChunkedReader splits a reader into several
// logical readers with a limited content size.
// Once the reader reaches its limit it provides
// an io.EOF.
// It can be continued by Calling Next, which returns
// whether a follow-up is required or not.
type ChunkedReader struct {
	lock      sync.Mutex
	reader    *LookAheadReader
	chunkSize int64
	chunkNo   int

	limited *io.LimitedReader
	err     error
}

var _ io.Reader = (*ChunkedReader)(nil)

func NewChunkedReader(r io.Reader, chunkSize int64) *ChunkedReader {
	return &ChunkedReader{
		reader:    NewLookAheadReader(r),
		chunkSize: chunkSize,
		limited:   &io.LimitedReader{r, chunkSize},
	}
}

func (c *ChunkedReader) Read(p []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	n, err = c.limited.Read(p)
	c.err = err
	return n, err
}

// ChunkNo returns the number previously
// provided chunks.
func (c *ChunkedReader) ChunkNo() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.chunkNo
}

// ChunkDone returns true, if the actual
// chunk is completely read.
func (c *ChunkedReader) ChunkDone() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return errors.Is(c.err, io.EOF)
}

// Next returns true, if a followup chunk
// has been prepared for the reader.
// If called while the current chunk is not yet completed
// it always returns false (check by calling ChunkDone).
func (c *ChunkedReader) Next() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !errors.Is(c.err, io.EOF) {
		return false
	}

	if c.limited.N > 0 {
		// don't need to check for more data if EOF is
		// provided before chunk size is reached.
		return false
	}
	// cannot assume that read with size 0 returns EOF as proposed
	// by io.Reader.Read (see bytes.Buffer.Read).
	// Therefore, we really have to read something.
	var buf [1]byte
	n, err := c.reader.LookAhead(buf[:])
	if n == 0 && errors.Is(err, io.EOF) {
		return false
	}

	c.chunkNo++
	c.err = nil
	c.limited = &io.LimitedReader{c.reader, c.chunkSize}
	return true
}
