package accessio

import (
	"bytes"
	"io"
	"sync"
)

// LookAheadReader is an io.Reader which additionally
// provides a look ahead of upcoming data, which does not
// affect the regular reader.
type LookAheadReader struct {
	lock   sync.Mutex
	reader io.Reader
	buffer *bytes.Buffer
}

func NewLookAheadReader(r io.Reader) *LookAheadReader {
	return &LookAheadReader{
		reader: r,
	}
}

func (r *LookAheadReader) Read(p []byte) (int, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.read(p)
}

func (r *LookAheadReader) read(p []byte) (int, error) {
	var (
		n   int
		err error
	)

	if r.buffer != nil && r.buffer.Len() > 0 {
		// first, consume from buffer
		n, err = r.buffer.Read(p)
		if err != nil { // the only error returned is io.EOF
			r.buffer = nil
		}
	} else {
		r.buffer = nil
	}

	if n >= len(p) {
		return n, nil
	}

	cnt, err := r.reader.Read(p[n:])
	return cnt + n, err
}

// LookAhead  provides a preview of upcoming data.
// It tries to fill the complete given buffer.
// The regular data stream provided by Read is not affected.
func (r *LookAheadReader) LookAhead(p []byte) (n int, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	cnt := 0
	for cnt < len(p) {
		n, err = r.read(p[cnt:])
		if err != nil {
			break
		}
		cnt += n
	}

	if cnt > 0 {
		if r.buffer == nil {
			r.buffer = bytes.NewBuffer(nil)
		}
		r.buffer.Write(p[:cnt])
	}

	return cnt, err
}
