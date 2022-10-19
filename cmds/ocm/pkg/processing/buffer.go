// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package processing

import (
	"fmt"
	"sync"

	"github.com/containerd/containerd/pkg/atomic"
	"github.com/mandelsoft/logging"

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

type BufferCreator func(log logging.Context) ProcessingBuffer

type ProcessingIterable interface {
	ProcessingIterator() ProcessingIterator
	Iterator() data.Iterator
}

type ProcessingIterator interface {
	HasNext() bool
	NextProcessingEntry() ProcessingEntry
}

type ProcessingBuffer interface {
	Add(e ProcessingEntry) ProcessingBuffer
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
	Add(e ProcessingEntry) bool
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
	log      logging.Context
}

var (
	_ ProcessingBuffer = &_buffer{}
	_ data.Iterable    = &_buffer{}
)

func NewProcessingBuffer(log logging.Context, i BufferImplementation) ProcessingBuffer {
	return (&_buffer{}).new(log, i)
}

func (this *_buffer) new(log logging.Context, i BufferImplementation) *_buffer {
	this.BufferImplementation = i
	this.Cond = sync.NewCond(&this.Mutex)
	this.complete = atomic.NewBool(false)
	i.SetFrame(this)
	this.log = log
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
	log     logging.Context
}

func NewSimpleBuffer(log logging.Context) ProcessingBuffer {
	return NewProcessingBuffer(log, (&simpleBuffer{}).new(log))
}

func (this *simpleBuffer) new(log logging.Context) *simpleBuffer {
	this.entries = []ProcessingEntry{}
	this.log = log
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
	return (&simpleBufferIterator{}).new(this, true, this.log)
}

func (this *simpleBuffer) ProcessingIterator() ProcessingIterator {
	return (&simpleBufferIterator{}).new(this, false, this.log)
}

func (this *simpleBuffer) Add(e ProcessingEntry) bool {
	this.entries = append(this.entries, e)
	return true
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
	log     logging.Context
}

var (
	_ ProcessingIterator = &simpleBufferIterator{}
	_ data.Iterator      = &simpleBufferIterator{}
)

func (this *simpleBufferIterator) new(buffer *simpleBuffer, valid bool, log logging.Context) *simpleBufferIterator {
	this.valid = valid
	this.current = -1
	this.buffer = buffer
	this.log = log
	return this
}

func (this *simpleBufferIterator) HasNext() bool {
	this.buffer.frame.Lock()
	defer this.buffer.frame.Unlock()
	for {
		this.log.Logger().Debug("HasNext", "current", this.current)
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
		this.log.Logger().Debug("NextProcessingEntry", "current", this.current)
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
// while the value iterator offers a sequence according the order
// of the initially specified indices.
type orderedBuffer struct {
	simple    simpleBuffer
	root      data.DLLRoot
	last      *data.DLL
	valid     *data.DLL
	nextIndex Index
	size      int
	log       logging.Context
}

type CheckNext interface {
	CheckNext() bool
}

func NewOrderedBuffer(log logging.Context) ProcessingBuffer {
	return NewProcessingBuffer(log, (&orderedBuffer{}).new(log))
}

func (this *orderedBuffer) new(log logging.Context) *orderedBuffer {
	(&this.simple).new(log)
	this.root.New(this)
	this.valid = this.root.DLL()
	this.last = this.valid
	this.nextIndex = this.nextIndex.Next(-1, 0)
	this.log = log
	return this
}

func (this *orderedBuffer) SetFrame(frame BufferFrame) {
	this.simple.SetFrame(frame)
}

func (this *orderedBuffer) Add(e ProcessingEntry) bool {
	e.Index.Validate(e.MaxIndex)
	this.simple.Add(e)
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
	c.Append(n)
	this.size++
	if n.Next() == nil {
		this.last = n
	}

	increased := false
	this.log.Logger().Debug("add index to cur value", "index", e.Index, "value", e.Value, "next-index", this.nextIndex)

	next := this.valid.Next()
	for next != nil && !next.Get().(*ProcessingEntry).Index.After(this.nextIndex) {
		n := next.Get().(*ProcessingEntry)
		this.nextIndex = n.Index.Next(n.MaxIndex, n.MaxSub)
		this.valid = next
		next = next.Next()
		increased = true
		this.log.Logger().Debug("increase to index to value", "index", n.Index, "value", n.Value)
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
			this.current = n // always proceed
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

func NewEntryIterableFromIterable(log logging.Context, data data.Iterable) ProcessingIterable {
	e, ok := data.(ProcessingIterable)
	if ok {
		return e
	}
	c := NewOrderedBuffer(log)

	go func() {
		i := data.Iterator()
		for idx := 0; i.HasNext(); idx++ {
			c.Add(ProcessingEntry{Top(idx), -1, 0, true, i.Next()})
		}
		c.Close()
	}()
	return c
}
