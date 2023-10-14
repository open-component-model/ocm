// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package refmgmt

import (
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/errors"
)

var ErrClosed = errors.ErrClosed()

// ReferencableCloser manages closable views to a basic closer.
// If the last view is closed, the basic closer is finally closed.
type ReferencableCloser interface {
	Allocatable

	RefCount() int
	UnrefLast() error
	IsClosed() bool

	Closer() io.Closer
	View(main ...bool) (CloserView, error)

	WithName(name string) ReferencableCloser
}

type referencableCloser struct {
	RefMgmt
	closer io.Closer
}

func NewRefCloser(closer io.Closer, unused ...bool) ReferencableCloser {
	return &referencableCloser{RefMgmt: NewAllocatable(closer.Close, unused...), closer: closer}
}

func (r *referencableCloser) WithName(name string) ReferencableCloser {
	r.RefMgmt.WithName(name)
	return r
}

func (r *referencableCloser) Closer() io.Closer {
	return r.closer
}

// View creates a new closable view.
// The object is closed if the last view has been released.
// If main is set to true, the close method of the view
// returns an error, if it is not the last reference.
func (r *referencableCloser) View(main ...bool) (CloserView, error) {
	err := r.Ref()
	if err != nil {
		return nil, err
	}
	v := &view{ref: r}
	for _, b := range main {
		if b {
			v.main = true
		}
	}
	return v, nil
}

type LazyMode interface {
	Lazy()
}

func Lazy(o interface{}) bool {
	if l, ok := o.(LazyMode); ok {
		l.Lazy()
		return true
	}
	return false
}

func CloseTemporary(c io.Closer) error {
	if !Lazy(c) {
		return errors.ErrNotSupported("lazy mode")
	}
	return c.Close()
}

func PropagateCloseTemporary(errp *error, c io.Closer) {
	errors.PropagateError(errp, func() error { return CloseTemporary(c) })
}

type CloserView interface {
	io.Closer
	LazyMode

	IsClosed() bool

	View() (CloserView, error)

	Release() error
	Finalize() error

	Closer() io.Closer

	Execute(f func() error) error
}

type view struct {
	lock   sync.Mutex
	ref    ReferencableCloser
	main   bool
	closed bool
}

var _ CloserView = (*view)(nil)

func (v *view) Lazy() {
	v.main = false
}

func (v *view) Execute(f func() error) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	if v.closed {
		return ErrClosed
	}
	return f()
}

func (v *view) Release() error {
	v.lock.Lock()
	defer v.lock.Unlock()
	if v.closed {
		return ErrClosed
	}
	v.closed = true
	return v.ref.Unref()
}

func (v *view) Finalize() error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.closed {
		return ErrClosed
	}

	if err := v.ref.UnrefLast(); err != nil {
		return errors.Wrapf(err, "unable to unref last")
	}
	v.closed = true
	return nil
}

func (v *view) Close() error {
	if v.main {
		return v.Finalize()
	}

	return v.Release()
}

func (v *view) IsClosed() bool {
	v.lock.Lock()
	defer v.lock.Unlock()

	return v.closed
}

func (v *view) View() (CloserView, error) {
	return v.ref.View()
}

func (v *view) Closer() io.Closer {
	return v.ref.Closer()
}

type Closers []io.Closer

func (c *Closers) Add(closers ...io.Closer) {
	for _, e := range closers {
		if e != nil {
			*c = append(*c, e)
		}
	}
}

func (c Closers) Effective() io.Closer {
	switch len(c) {
	case 0:
		return nil
	case 1:
		return c[0]
	default:
		return c
	}
}

func (c Closers) Close() error {
	list := errors.ErrList()
	for _, e := range c {
		list.Add(e.Close())
	}
	return list.Result()
}

type CloserFunc func() error

func (c CloserFunc) Close() error {
	return c()
}
