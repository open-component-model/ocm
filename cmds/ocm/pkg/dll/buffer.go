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

package dll

import (
	_ "fmt"
	"sync"

	"github.com/containerd/containerd/pkg/atomic"
)

type Index int

type ProcessingEntry struct {
	Index Index
	Valid bool
	Value interface{}
}

func NewEntry(i Index, v interface{}, valid ...bool) ProcessingEntry {
	return ProcessingEntry{
		Index: i,
		Valid: len(valid) == 0 || valid[0],
		Value: v,
	}
}

type BufferCreator func() ProcessingBuffer

type ProcessingIterable interface {
	ProcessingIterator() ProcessingIterator
	Iterator() Iterator
}

type ProcessingIterator interface {
	HasNext() bool
	NextProcessingEntry() ProcessingEntry
}

type ProcessingBuffer interface {
	Add(e ProcessingEntry) ProcessingBuffer
	Open()
	Close()

	ProcessingIterable

	IsClosed() bool
}

type BufferFrame interface {
	Lock()
	Unlock()
	Broadcast()
	Wait()

	IsClosed() bool
}

type BufferImplementation interface {
	Add(e ProcessingEntry) bool
	Open()
	Close()

	ProcessingIterable

	SetFrame(frame BufferFrame)
}

type _buffer struct {
	BufferImplementation
	*sync.Cond
	sync.Mutex
	complete atomic.Bool
}

var _ ProcessingBuffer = &_buffer{}
var _ Iterable = &_buffer{}

func NewProcessingBuffer(i BufferImplementation) ProcessingBuffer {
	return (&_buffer{}).new(i)
}

func (this *_buffer) new(i BufferImplementation) *_buffer {
	this.BufferImplementation = i
	this.Cond = sync.NewCond(&this.Mutex)
	this.complete = atomic.NewBool(false)
	i.SetFrame(this)
	return this
}

func (this *_buffer) Add(e ProcessingEntry) ProcessingBuffer {
	this.Lock()
	notify := this.BufferImplementation.Add(e)
	this.Unlock()
	if notify {
		this.Broadcast()
	}
	return this
}

func (this *_buffer) Open() {
	this.Lock()
	this.BufferImplementation.Open()
	this.complete.Unset()
	this.Unlock()
}

func (this *_buffer) Close() {
	this.Lock()
	this.BufferImplementation.Close()
	this.complete.Set()
	this.Unlock()
	this.Broadcast()
}

func (this *_buffer) IsClosed() bool {
	return this.complete.IsSet()
}

////////////////////////////////////////////////////////////////////////////////

type simpleBuffer struct {
	frame   BufferFrame
	entries []ProcessingEntry
}

func NewSimpleBuffer() ProcessingBuffer {
	return NewProcessingBuffer((&simpleBuffer{}).new())
}

func (this *simpleBuffer) new() *simpleBuffer {
	this.entries = []ProcessingEntry{}
	return this
}

func (this *simpleBuffer) SetFrame(frame BufferFrame) {
	this.frame = frame
}

func (this *simpleBuffer) Open() {
}

func (this *simpleBuffer) Close() {
}

func (this *simpleBuffer) Iterator() Iterator {
	return (&simpleBufferIterator{}).new(this, true)
}

func (this *simpleBuffer) ProcessingIterator() ProcessingIterator {
	return (&simpleBufferIterator{}).new(this, false)
}

func (this *simpleBuffer) Add(e ProcessingEntry) bool {
	this.entries = append(this.entries, e)
	return true
}

type simpleBufferIterator struct {
	buffer  *simpleBuffer
	valid   bool
	current int
}

var _ ProcessingIterator = &simpleBufferIterator{}
var _ Iterator = &simpleBufferIterator{}

func (this *simpleBufferIterator) new(buffer *simpleBuffer, valid bool) *simpleBufferIterator {
	this.valid = valid
	this.current = -1
	this.buffer = buffer
	return this
}

func (this *simpleBufferIterator) HasNext() bool {
	this.buffer.frame.Lock()
	defer this.buffer.frame.Unlock()
	for true {
		//fmt.Printf("HasNext: %d(%d) %t\n", this.current, this.container.len(), this.container.closed())
		if len(this.buffer.entries) > this.current+1 {
			if !this.valid || this.buffer.entries[this.current+1].Valid {
				return true
			}
			this.current++
			continue
		}
		if this.buffer.frame.IsClosed() {
			return false
		}
		this.buffer.frame.Wait()
	}
	return false
}

func (this *simpleBufferIterator) Next() interface{} {
	return this.NextProcessingEntry().Value
}

func (this *simpleBufferIterator) NextProcessingEntry() ProcessingEntry {
	this.buffer.frame.Lock()
	defer this.buffer.frame.Unlock()
	for {
		//fmt.Printf("HasNext: %d(%d) %t\n", this.current, this.container.len(), this.container.closed())
		if len(this.buffer.entries) > this.current+1 {
			this.current++
			if !this.valid || this.buffer.entries[this.current].Valid {
				return this.buffer.entries[this.current]
			}
			continue
		}
		if this.buffer.frame.IsClosed() {
			return ProcessingEntry{}
		}
		this.buffer.frame.Wait()
	}
}

////////////////////////////////////////////////////////////////////////////

// orderedBuffer is a buffer view offering an ordered list of entries.
// the entry iterator provides access to an unordered sequence
// while the value iterator offeres a sequence according the order
// of the initially specified indices
type orderedBuffer struct {
	simple    simpleBuffer
	root      DLLRoot
	last      *DLL
	valid     *DLL
	nextIndex Index
}

type orderedEntry struct {
	dll   DLL
	entry ProcessingEntry
}

func NewOrderedBuffer() ProcessingBuffer {
	return NewProcessingBuffer((&orderedBuffer{}).new())
}

func (this *orderedBuffer) new() *orderedBuffer {
	(&this.simple).new()
	this.root.New(this)
	this.valid = this.root.DLL()
	this.last = this.valid
	return this
}

func (this *orderedBuffer) SetFrame(frame BufferFrame) {
	this.simple.SetFrame(frame)
}

func (this *orderedBuffer) Add(e ProcessingEntry) bool {
	this.simple.Add(e)
	n := NewDLL(&e)

	c := this.root.DLL()
	i := c.Next()
	for i != nil {
		v := i.Get().(*ProcessingEntry)
		if v.Index > e.Index {
			break
		}
		c, i = i, i.Next()
	}
	c.Append(n)
	if n.Next() == nil {
		this.last = n
	}

	increased := false
	next := this.valid.Next()
	for next != nil && next.Get().(*ProcessingEntry).Index <= this.nextIndex {
		this.nextIndex = next.Get().(*ProcessingEntry).Index + 1
		this.valid = next
		next = next.Next()
		increased = true
	}
	return increased
}

func (this *orderedBuffer) Close() {
	this.simple.Close()
	if this.valid != this.last {
		this.valid = this.last
		this.nextIndex = this.valid.Get().(*ProcessingEntry).Index
	}
}

func (this *orderedBuffer) Open() {
	this.simple.Open()
}

func (this *orderedBuffer) Iterator() Iterator {
	// this this is another this than this in iter() in this.container
	// still inherited to offer the unordered entries for processing
	return (&orderedBufferIterator{}).new(this)
}

func (this *orderedBuffer) ProcessingIterator() ProcessingIterator {
	return this.simple.ProcessingIterator()
}

type orderedBufferIterator struct {
	buffer  *orderedBuffer
	current *DLL
}

var _ Iterator = (*orderedBufferIterator)(nil)

func (this *orderedBufferIterator) new(buffer *orderedBuffer) *orderedBufferIterator {
	this.buffer = buffer
	this.current = this.buffer.root.DLL()
	return this
}

func (this *orderedBufferIterator) HasNext() bool {
	this.buffer.simple.frame.Lock()
	defer this.buffer.simple.frame.Unlock()
	for {
		n := this.current.Next()
		if n != nil && this.current != this.buffer.valid {
			if n.Get().(*ProcessingEntry).Valid {
				return true
			}
			this.current = n // skip invalid entries
			continue
		}
		if this.buffer.simple.frame.IsClosed() {
			return false
		}
		this.buffer.simple.frame.Wait()
	}
}

func (this *orderedBufferIterator) Next() interface{} {
	this.buffer.simple.frame.Lock()
	defer this.buffer.simple.frame.Unlock()
	for {
		n := this.current.Next()
		if n != nil && this.current != this.buffer.valid {
			e := n.Get().(*ProcessingEntry)
			this.current = n //always proceed
			if e.Valid {
				return e.Value
			}
			continue
		}
		if this.buffer.simple.frame.IsClosed() {
			return ProcessingEntry{}
		}
		this.buffer.simple.frame.Wait()
	}
}

////////////////////////////////////////////////////////////////////////////

type valueIterator struct {
	ProcessingIterator
}

func (i *valueIterator) Next() interface{} {
	return i.NextProcessingEntry().Value
}

func ValueIterator(i ProcessingIterator) Iterator {
	return &valueIterator{i}
}

type valueIterable struct {
	ProcessingIterable
}

func (i *valueIterable) Iterator() Iterator {
	return ValueIterator(i.ProcessingIterator())
}

func ValueIterable(i ProcessingIterable) ProcessingIterable {
	return &valueIterable{i}
}

func NewEntryIterableFromIterable(data Iterable) ProcessingIterable {
	e, ok := data.(ProcessingIterable)
	if ok {
		return e
	}
	c := NewOrderedBuffer()

	go func() {

		i := data.Iterator()
		for idx := 0; i.HasNext(); idx++ {
			c.Add(ProcessingEntry{Index(idx), true, i.Next()})
		}
		c.Close()
	}()
	return c
}
