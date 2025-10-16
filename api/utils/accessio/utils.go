package accessio

import (
	"fmt"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/utils/compression"
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
	return compression.NopWriteCloser(w)
}

////////////////////////////////////////////////////////////////////////////////

type once struct {
	callbacks []CloserCallback
	closer    io.Closer
}

type CloserCallback func()

func OnceCloser(c io.Closer, callbacks ...CloserCallback) io.Closer {
	return &once{callbacks, c}
}

func (c *once) Close() error {
	if c.closer == nil {
		return nil
	}

	t := c.closer
	c.closer = nil
	err := t.Close()

	for _, cb := range c.callbacks {
		cb()
	}

	if err != nil {
		return fmt.Errorf("unable to close: %w", err)
	}

	return nil
}

func Close(closer ...io.Closer) error {
	if len(closer) == 0 {
		return nil
	}
	list := errors.ErrList()
	for _, c := range closer {
		if c != nil {
			list.Add(c.Close())
		}
	}
	return list.Result()
}

type Closer func() error

func (c Closer) Close() error {
	return c()
}
