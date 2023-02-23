// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/utils/panics"
)

type DLLRoot struct {
	lock sync.RWMutex
	root DLL
}

func (this *DLLRoot) New(p interface{}) *DLLRoot {
	if this == nil {
		this = &DLLRoot{}
	}
	if p == nil {
		p = this
	}
	this.root.payload = p
	this.root.lock = &this.lock
	return this
}

func (this *DLLRoot) Append(d *DLL) {
	this.root.Append(d)
}

func (this *DLLRoot) Next() *DLL {
	return this.root.Next()
}

func (this *DLLRoot) DLL() *DLL {
	return &this.root
}

func (this *DLLRoot) Iterator() Iterator {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return &dllIterator{
		lock:    &this.lock,
		current: &this.root,
	}
}

type DLL struct {
	lock    *sync.RWMutex
	prev    *DLL
	next    *DLL
	payload interface{}
}

func NewDLL(p interface{}) *DLL {
	return &DLL{payload: p}
}

func (this *DLL) Next() *DLL {
	return this.next
}

func (this *DLL) Prev() *DLL {
	if this.prev == nil {
		return nil
	}
	return this.prev
}

func (this *DLL) Get() interface{} {
	return this.payload
}

func (this *DLL) Set(p interface{}) {
	this.payload = p
}

func (this *DLL) Append(d *DLL) {
	defer panics.HandlePanic()
	if d.next != nil || d.prev != nil {
		panic("dll element already in use")
	}
	if this.lock != nil {
		this.lock.Lock()
		defer this.lock.Unlock()
		d.lock = this.lock
	}
	d.next = this.next
	d.prev = this
	if this.next != nil {
		this.next.prev = d
	}
	this.next = d
}

func (this *DLL) Remove() {
	if this.prev != nil {
		this.prev.next = this.next
	}
	if this.next != nil {
		this.next.prev = this.prev
	}
	this.next = nil
	this.prev = nil
}

////////////////////////////////////////////////////////////////////////////////

type dllIterator struct {
	lock    *sync.RWMutex
	current *DLL
}

var _ Iterator = (*dllIterator)(nil)

func (this *dllIterator) _lock() {
	if this.lock != nil {
		this.lock.RLock()
	}
}

func (this *dllIterator) _unlock() {
	if this.lock != nil {
		this.lock.RUnlock()
	}
}

func (this *dllIterator) HasNext() bool {
	this._lock()
	defer this._unlock()
	return this.current.next != nil
}

func (this *dllIterator) Next() interface{} {
	this._lock()
	defer this._unlock()
	if this.current.next != nil {
		this.current = this.current.next
		return this.current
	}
	return nil
}

func (this *dllIterator) NextDLL() *DLL {
	this._lock()
	defer this._unlock()
	if this.current.next != nil {
		this.current = this.current.next
		return this.current
	}
	return nil
}
