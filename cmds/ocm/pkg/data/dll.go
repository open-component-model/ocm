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

package data

import (
	"sync"
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
