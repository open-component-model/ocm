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

package processing

import (
	"fmt"
	_ "fmt"
	"sync"

	"github.com/containerd/containerd/pkg/atomic"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
)

type Index = IndexArray

func Top(i int) Index {
	return IndexArray{i}
}

type IndexArray []int

func (i IndexArray) After(o IndexArray) bool {
	for l, v := range i {
		if l >= len(o) {
			return true
		}
		if v != o[l] {
			return v > o[l]
		}
	}
	return len(i) > len(o)
}

func (i IndexArray) Next(max, sub int) IndexArray {
	l := len(i)
	n := i.Copy()

	if sub > 0 || len(i) == 0 {
		return append(n, 0)
	}
	n[l-1]++
	if max > 0 && n[l-1] >= max {
		n[l-2]++
		return n[:l-1]
	}
	return n
}

func (i IndexArray) Copy() IndexArray {
	n := make(IndexArray, len(i))
	copy(n, i)
	return n
}

func (i IndexArray) Validate(max int) {
	if max >= 0 && i[len(i)-1] >= max {
		panic(fmt.Sprintf("index %d >= max %d", i[len(i)-1], max))
	}
}

type ProcessingEntry struct {
	Index    Index
	MaxIndex int
	MaxSub   int
	Valid    bool
	Value    interface{}
}

type SubEntries int

func NewEntry(i Index, v interface{}, opts ...interface{}) ProcessingEntry {
	max := -1
	sub := 0
	valid := true
	for _, o := range opts {
		switch t := o.(type) {
		case bool:
			valid = valid && t
		case SubEntries:
			sub = int(t)
		case int:
			max = t
		default:
			panic(fmt.Errorf("invalid entry option %T", o))
		}
	}
	if len(i) > 1 && max < 0 {
		panic(fmt.Errorf("invalid max option %d", max))
	}
	return ProcessingEntry{
		Index:    i,
		Valid:    valid,
		Value:    v,
		MaxIndex: max,
		MaxSub:   sub,
	}
}

type BufferCreator func() ProcessingBuffer

type ProcessingIterable interface {
	ProcessingIterator() ProcessingIterator
	Iterator() data.Iterator
}

type ProcessingIterator interface {
	HasNext() bool
	NextProcessingEntry() ProcessingEntry
}

type ProcessingBuffer interface {
	Add(e ProcessingEntry) (ProcessingBuffer, error)
	Len() int
	Get(int) interface{}
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
	Add(e ProcessingEntry) (bool, error)
	Open()
	Close()
	Len() int
	Get(i int) interface{}

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
var _ data.Iterable = &_buffer{}

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

func (this *_buffer) Add(e ProcessingEntry) (ProcessingBuffer, error) {
	this.Lock()
	notify, err := this.BufferImplementation.Add(e)
	if err != nil {
		return nil, fmt.Errorf("buffer add failed: %w", err)
	}
	this.Unlock()
	if notify {
		this.Broadcast()
	}
	return this, nil
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

func (this *_buffer) Len() int {
	this.Lock()
	defer this.Unlock()
	return this.BufferImplementation.Len()
}

func (this *_buffer) Get(i int) interface{} {
	this.Lock()
	defer this.Unlock()
	return this.BufferImplementation.Get(i)
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

func (this *simpleBuffer) Iterator() data.Iterator {
	return (&simpleBufferIterator{}).new(this, true)
}

func (this *simpleBuffer) ProcessingIterator() ProcessingIterator {
	return (&simpleBufferIterator{}).new(this, false)
}

func (this *simpleBuffer) Add(e ProcessingEntry) (bool, error) {
	this.entries = append(this.entries, e)
	return true, nil
}

func (this *simpleBuffer) Len() int {
	return len(this.entries)
}

func (this *simpleBuffer) Get(i int) interface{} {
	e := this.entries[i]
	if e.Valid {
		return e.Value
	}
	return nil
}

type simpleBufferIterator struct {
	buffer  *simpleBuffer
	valid   bool
	current int
}

var _ ProcessingIterator = &simpleBufferIterator{}
var _ data.Iterator = &simpleBufferIterator{}

func (this *simpleBufferIterator) new(buffer *simpleBuffer, valid bool) *simpleBufferIterator {
	this.valid = valid
	this.current = -1
	this.buffer = buffer
	return this
}

func (this *simpleBufferIterator) HasNext() bool {
	this.buffer.frame.Lock()
	defer this.buffer.frame.Unlock()
	for {
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
	root      data.DLLRoot
	last      *data.DLL
	valid     *data.DLL
	nextIndex Index
	size      int
}

type CheckNext interface {
	CheckNext() bool
}

func NewOrderedBuffer() ProcessingBuffer {
	return NewProcessingBuffer((&orderedBuffer{}).new())
}

func (this *orderedBuffer) new() *orderedBuffer {
	(&this.simple).new()
	this.root.New(this)
	this.valid = this.root.DLL()
	this.last = this.valid
	this.nextIndex = this.nextIndex.Next(-1, 0)
	return this
}

func (this *orderedBuffer) SetFrame(frame BufferFrame) {
	this.simple.SetFrame(frame)
}

func (this *orderedBuffer) Add(e ProcessingEntry) (bool, error) {
	e.Index.Validate(e.MaxIndex)
	if _, err := this.simple.Add(e); err != nil {
		return false, fmt.Errorf("ordered buffer add failed: %w", err)
	}
	n := data.NewDLL(&e)

	c := this.root.DLL()
	i := c.Next()
	for i != nil {
		v := i.Get().(*ProcessingEntry)
		if v.Index.After(e.Index) {
			break
		}
		c, i = i, i.Next()
	}
	if err := c.Append(n); err != nil {
		return false, fmt.Errorf("ordered buffer add failed: %w", err)
	}
	this.size++
	if n.Next() == nil {
		this.last = n
	}

	increased := false
	next := this.valid.Next()
	for next != nil && !next.Get().(*ProcessingEntry).Index.After(this.nextIndex) {
		n := next.Get().(*ProcessingEntry)
		this.nextIndex = n.Index.Next(n.MaxIndex, n.MaxSub)
		this.valid = next
		next = next.Next()
		increased = true
	}
	return increased, nil
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

func (this *orderedBuffer) Iterator() data.Iterator {
	// this this is another this than this in iter() in this.container
	// still inherited to offer the unordered entries for processing
	return (&orderedBufferIterator{}).new(this)
}

func (this *orderedBuffer) ProcessingIterator() ProcessingIterator {
	return this.simple.ProcessingIterator()
}

func (this *orderedBuffer) Len() int {
	return this.size
}

func (this *orderedBuffer) Get(i int) interface{} {
	e := this.root.DLL()
	for e != nil && i >= 0 {
		e = e.Next()
		i--
	}
	if e == nil {
		return nil
	}
	pe := e.Get().(*ProcessingEntry)
	if pe.Valid {
		return pe.Value
	}
	return nil
}

type orderedBufferIterator struct {
	buffer  *orderedBuffer
	current *data.DLL
}

var _ data.Iterator = (*orderedBufferIterator)(nil)

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

func (this *orderedBufferIterator) CheckNext() bool {
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
		return false
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

func ValueIterator(i ProcessingIterator) data.Iterator {
	return &valueIterator{i}
}

type valueIterable struct {
	ProcessingIterable
}

func (i *valueIterable) Iterator() data.Iterator {
	return ValueIterator(i.ProcessingIterator())
}

func ValueIterable(i ProcessingIterable) ProcessingIterable {
	return &valueIterable{i}
}

func NewEntryIterableFromIterable(data data.Iterable) ProcessingIterable {
	e, ok := data.(ProcessingIterable)
	if ok {
		return e
	}
	c := NewOrderedBuffer()

	go func() {
		i := data.Iterator()
		for idx := 0; i.HasNext(); idx++ {
			if _, err := c.Add(ProcessingEntry{Top(idx), -1, 0, true, i.Next()}); err != nil {
				fmt.Printf("failed to add entries: %s", err)
				break
			}
		}
		c.Close()
	}()
	return c
}
