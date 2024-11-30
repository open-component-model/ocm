package ioutils

import (
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/mandelsoft/goutils/general"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/sliceutils"
)

////////////////////////////////////////////////////////////////////////////////

type additionalCloser[T any] struct {
	msg              []interface{}
	wrapped          T
	additionalCloser io.Closer
}

func (c *additionalCloser[T]) Close() error {
	var list *errors.ErrorList
	if len(c.msg) == 0 {
		list = errors.ErrListf("close")
	} else {
		if s, ok := c.msg[0].(string); ok && len(c.msg) > 1 {
			list = errors.ErrListf(s, c.msg[1:]...)
		} else {
			list = errors.ErrList(c.msg...)
		}
	}
	if cl, ok := generics.TryCast[io.Closer](c.wrapped); ok {
		list.Add(cl.Close())
	}
	if c.additionalCloser != nil {
		list.Add(c.additionalCloser.Close())
	}
	return list.Result()
}

func newAdditionalCloser[T any](w T, closer io.Closer, msg ...interface{}) additionalCloser[T] {
	return additionalCloser[T]{
		wrapped:          w,
		msg:              msg,
		additionalCloser: closer,
	}
}

////////////////////////////////////////////////////////////////////////////////

type readCloser struct {
	additionalCloser[io.Reader]
}

var _ io.ReadCloser = (*readCloser)(nil)

// Deprecated: use AddReaderCloser .
func AddCloser(reader io.ReadCloser, closer io.Closer, msg ...string) io.ReadCloser {
	return AddReaderCloser(reader, closer, sliceutils.AsAny(msg)...)
}

func ReadCloser(r io.Reader) io.ReadCloser {
	return AddReaderCloser(r, nil)
}

func AddReaderCloser(reader io.Reader, closer io.Closer, msg ...interface{}) io.ReadCloser {
	return &readCloser{
		additionalCloser: newAdditionalCloser[io.Reader](reader, closer, msg...),
	}
}

func (c *readCloser) Read(p []byte) (n int, err error) {
	return c.wrapped.Read(p)
}

type writeCloser struct {
	additionalCloser[io.Writer]
}

var _ io.WriteCloser = (*writeCloser)(nil)

func WriteCloser(w io.Writer) io.WriteCloser {
	return AddWriterCloser(w, nil)
}

func AddWriterCloser(writer io.Writer, closer io.Closer, msg ...interface{}) io.WriteCloser {
	return &writeCloser{
		additionalCloser: newAdditionalCloser[io.Writer](writer, closer, msg...),
	}
}

func (c *writeCloser) Write(p []byte) (n int, err error) {
	return c.wrapped.Write(p)
}

////////////////////////////////////////////////////////////////////////////////

type DupReadCloser interface {
	io.ReadCloser
	Dup() (DupReadCloser, error)
}

// dupReadCloser is the internal representation
// with ref counting for a dup reader view.
type dupReadCloser struct {
	rc    io.ReadCloser
	count atomic.Int64
}

var _ DupReadCloser = (*dupViewReadCloser)(nil)

type dupViewReadCloser struct {
	lock sync.Mutex
	r    *dupReadCloser
}

func (d *dupViewReadCloser) Read(p []byte) (n int, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.r == nil {
		return 0, os.ErrClosed
	}
	return d.r.Read(p)
}

func (d *dupViewReadCloser) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.r == nil {
		return os.ErrClosed
	}
	err := d.r.Close()
	d.r = nil
	return err
}

func (d *dupViewReadCloser) Dup() (DupReadCloser, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.r == nil {
		return nil, os.ErrClosed
	}
	return d.r.Dup(), nil
}

////

func (d *dupReadCloser) Read(p []byte) (n int, err error) {
	return d.rc.Read(p)
}

func (d *dupReadCloser) Close() error {
	c := d.count.Add(-1)
	if c < 0 {
		return os.ErrClosed
	}
	if c > 0 {
		return nil
	}
	return d.rc.Close()
}

func (d *dupReadCloser) Dup() DupReadCloser {
	d.count.Add(1)
	return &dupViewReadCloser{
		r: d,
	}
}

// NewDupReadCloser provides a reader which can be duplicated.
// Duplicated means, that a new separately closeable reader is
// provided. The original reader is closed with the close
// of the last provided reader view.
// The passed reader must never be explicitly closed.
// It is closed with the last provided DupReadCloser.
// If called for a DupReadCloser, just a new view is provided.
// If called for an already closed reader, read calls
// will provide the behaviour of the passed reader.
// Close will succeed until the view is closed, which
// will provide the behaviour of the passed reader.
func NewDupReadCloser(rc io.ReadCloser, errs ...error) (DupReadCloser, error) {
	if err := general.Optional(errs...); err != nil {
		return nil, err
	}
	if d, ok := rc.(DupReadCloser); ok {
		return d, nil
	}
	return (&dupReadCloser{
		rc: rc,
	}).Dup(), nil
}
