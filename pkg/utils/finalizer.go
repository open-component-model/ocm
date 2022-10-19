// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/errors"
)

// Finalizer gathers finalization functions and calls
// them by calling the Finalize method.
// Add and Finalize may be called in any sequence and number.
// Finalize just calls the aggregated functions between its
// last and the actual call.
// This way it can be used together with defer to clean up
// stuff when leaving a function and combine it with
// controlled intermediate cleanup need, for example as part of
// a loop block.
type Finalizer struct {
	lock    sync.Mutex
	pending []func() error
}

func (f *Finalizer) With(fi func() error) {
	if fi != nil {
		f.lock.Lock()
		defer f.lock.Unlock()

		f.pending = append(f.pending, fi)
	}
}

func (f *Finalizer) Close(c io.Closer) {
	if c != nil {
		f.With(c.Close)
	}
}

func (f *Finalizer) Include(c *Finalizer) {
	if c != nil {
		f.With(c.Finalize)
	}
}

func (f *Finalizer) Length() int {
	f.lock.Lock()
	defer f.lock.Unlock()
	return len(f.pending)
}

func (f *Finalizer) Finalize() error {
	f.lock.Lock()
	defer f.lock.Unlock()

	list := errors.ErrListf("finalize")
	l := len(f.pending)
	for i := range f.pending {
		list.Add(f.pending[l-i-1]())
	}
	f.pending = nil
	return list.Result()
}
