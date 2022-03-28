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
	_ "fmt"
	"sync"
)

type container_creator func() container

type entry_iterable interface {
	entry_iterator() entry_iterator
	Iterator() Iterator
}

type entry_iterator interface {
	Iterator
	next() entry
}

type container interface {
	Lock()
	Unlock()
	Broadcast()
	Wait()

	add(e entry) container
	len() int
	get(int) entry
	open()
	close()
	closed() bool

	entry_iterable
}

type _container_impl interface {
	_open()
	_close()
	_add(e entry) bool
}

type _container struct {
	self _container_impl
	*sync.Cond
	sync.Mutex
	entries  []entry
	complete bool
}

var _ container = &_container{}
var _ Iterable = &_container{}

func NewContainer() container {
	return (&_container{}).new()
}

func (this *_container) new() *_container {
	this.self = this
	this.entries = []entry{}
	this.Cond = sync.NewCond(&this.Mutex)
	return this
}

func (this *_container) _open() {
	this.complete = false
}
func (this *_container) _close() {
	this.complete = true
}
func (this *_container) _add(e entry) bool {
	this.entries = append(this.entries, e)
	return true
}

func (this *_container) Iterator() Iterator {
	return this.entry_iterator()
}

func (this *_container) add(e entry) container {
	this.Lock()
	notify := this.self._add(e)
	this.Unlock()
	if notify {
		this.Broadcast()
	}
	return this
}

func (this *_container) open() {
	this.Lock()
	this.self._open()
	this.Unlock()
}
func (this *_container) close() {
	this.Lock()
	this.self._close()
	this.Unlock()
	this.Broadcast()
}

func (this *_container) len() int {
	return len(this.entries)
}

func (this *_container) get(i int) entry {
	return this.entries[i]
}

func (this *_container) closed() bool {
	return this.complete
}

func (this *_container) entry_iterator() entry_iterator {
	return (&_entry_iterator{}).new(this)
}

type _entry_iterator struct {
	container container
	current   int
}

var _ entry_iterator = &_entry_iterator{}
var _ Iterator = &_entry_iterator{}

func (this *_entry_iterator) new(container container) entry_iterator {
	this.current = -1
	this.container = container
	return this
}

func (this *_entry_iterator) HasNext() bool {
	this.container.Lock()
	defer this.container.Unlock()
	for true {
		//fmt.Printf("HasNext: %d(%d) %t\n", this.current, this.container.len(), this.container.closed())
		if this.container.len() > this.current+1 {
			if this.container.get(this.current + 1).ok {
				return true
			}
			this.current++
			continue
		}
		if this.container.closed() {
			return false
		}
		this.container.Wait()
	}
	return false
}

func (this *_entry_iterator) Next() interface{} {
	n, ok := this._next()
	if !ok {
		return nil
	}
	return n.value
}

func (this *_entry_iterator) next() entry {
	n, ok := this._next()
	if !ok {
		panic("iteration beynond data structure")
	}
	return n
}

func (this *_entry_iterator) _next() (entry, bool) {
	var n entry
	ok := true
	this.container.Lock()
	if this.container.len() > this.current+1 {
		this.current++
		n = this.container.get(this.current)
	} else {
		ok = false
	}
	this.container.Unlock()
	return n, ok

}

////////////////////////////////////////////////////////////////////////////

// a container view offering an ordered list of entries.
// the container api (besides the raw entry iterator iter())
// features the actually valid ordered entry set.
// there ordered array may contain holes, depending on the
// processing order, therefore a dedicted valid attribute
// is maintained, indicating the actually complete part
// of the list

type ordered_container struct {
	_container
	ordered []*entry
	valid   int
}

func NewOrderedContainer() container {
	return (&ordered_container{}).new()
}

func (this *ordered_container) new() *ordered_container {
	this._container.new()
	this._container.self = this
	this.ordered = []*entry{}
	return this
}

func (this *ordered_container) _add(e entry) bool {
	if e.ok {
		this._container._add(e)
		if e.index >= len(this.ordered) {
			t := make([]*entry, e.index+1)
			copy(t, this.ordered)
			this.ordered = t
		}
		this.ordered[e.index] = &e
		for this.valid < len(this.ordered) && this.ordered[this.valid] != nil {
			this.valid++
		}
	}
	return e.ok
}

func (this *ordered_container) _close() {
	this._container._close()
	this.valid = len(this.ordered)
}

func (this *ordered_container) len() int {
	return this.valid
}

func (this *ordered_container) get(i int) entry {
	e := this.ordered[i]
	if e != nil {
		return *e
	}
	return entry{i, false, nil}
}

func (this *ordered_container) Iterator() Iterator {
	// this this is another this than this in iter() in this.container
	// still inherited to offer the unordered entries for processing
	return (&_entry_iterator{}).new(this)
}

////////////////////////////////////////////////////////////////////////////

type entry struct {
	index int
	ok    bool
	value interface{}
}

////////////////////////////////////////////////////////////////////////////

func newEntryIterableFromIterable(data Iterable) entry_iterable {
	e, ok := data.(entry_iterable)
	if ok {
		return e
	}
	c := NewOrderedContainer()

	go func() {

		i := data.Iterator()
		for idx := 0; i.HasNext(); idx++ {
			c.add(entry{idx, true, i.Next()})
		}
		c.close()
	}()
	return c
}

