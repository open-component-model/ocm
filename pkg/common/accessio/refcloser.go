// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package accessio

import (
	"io"
	"sync"
)

// ReferencableCloser manages closable views to a basic closer.
// If the last view is closed, the basic closer is finally closed.
type ReferencableCloser interface {
	RefMgmt

	Closer() io.Closer
	View(main ...bool) (CloserView, error)
}

type referencableCloser struct {
	RefMgmt
	closer io.Closer
}

func NewRefCloser(closer io.Closer, unused ...bool) ReferencableCloser {
	return &referencableCloser{RefMgmt: NewAllocatable(closer.Close, unused...), closer: closer}
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

type CloserView interface {
	io.Closer
	IsClosed() bool

	View() (CloserView, error)

	Release() error
	Finalize() error

	Closer() io.Closer
}

type view struct {
	lock   sync.Mutex
	ref    ReferencableCloser
	main   bool
	closed bool
}

var _ CloserView = (*view)(nil)

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
	err := v.ref.UnrefLast()
	if err == nil {
		v.closed = true
	}
	return err
}

func (v *view) Close() error {
	if v.main {
		return v.Release()
	}
	return v.Finalize()
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
