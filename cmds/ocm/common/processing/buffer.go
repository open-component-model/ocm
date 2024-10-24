package processing

import (
	"fmt"
	"sync"

	"github.com/containerd/containerd/pkg/atomic"
	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/utils/panics"
	"ocm.software/ocm/cmds/ocm/common/data"
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

func (i IndexArray) Next(maxIndex, sub int) IndexArray {
	l := len(i)
	n := i.Copy()

	if sub > 0 || len(i) == 0 {
		return append(n, 0)
	}
	n[l-1]++
	if maxIndex > 0 && n[l-1] >= maxIndex {
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

func (i IndexArray) Validate(maxIndex int) {
	if maxIndex >= 0 && i[len(i)-1] >= maxIndex {
		panic(fmt.Sprintf("index %d >= max %d", i[len(i)-1], maxIndex))
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
	// If this is caught the only upstream problem would be an empty entry.
	// Which is fine if the user understands that it can happen.
	defer panics.HandlePanic()

	maxOptions := -1
	sub := 0
	valid := true
	for _, o := range opts {
		switch t := o.(type) {
		case bool:
			valid = valid && t
		case SubEntries:
			sub = int(t)
		case int:
			maxOptions = t
		default:
			panic(fmt.Errorf("invalid entry option %T", o))
		}
	}
	if len(i) > 1 && maxOptions < 0 {
		panic(fmt.Errorf("invalid max option %d", maxOptions))
	}
	return ProcessingEntry{
		Index:    i,
		Valid:    valid,
		Value:    v,
		MaxIndex: maxOptions,
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

func (b *_buffer) new(log logging.Context, i BufferImplementation) *_buffer {
	b.BufferImplementation = i
	b.Cond = sync.NewCond(&b.Mutex)
	b.complete = atomic.NewBool(false)
	i.SetFrame(b)
	b.log = log
	return b
}

func (b *_buffer) Add(e ProcessingEntry) ProcessingBuffer {
	b.Lock()
	notify := b.BufferImplementation.Add(e)
	b.Unlock()
	if notify {
		b.Broadcast()
	}
	return b
}

func (b *_buffer) Open() {
	b.Lock()
	b.BufferImplementation.Open()
	b.complete.Unset()
	b.Unlock()
}

func (b *_buffer) Close() {
	b.Lock()
	b.BufferImplementation.Close()
	b.complete.Set()
	b.Unlock()
	b.Broadcast()
}

func (b *_buffer) IsClosed() bool {
	return b.complete.IsSet()
}

func (b *_buffer) Len() int {
	b.Lock()
	defer b.Unlock()
	return b.BufferImplementation.Len()
}

func (b *_buffer) Get(i int) interface{} {
	b.Lock()
	defer b.Unlock()
	return b.BufferImplementation.Get(i)
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

func (sb *simpleBuffer) new(log logging.Context) *simpleBuffer {
	sb.entries = []ProcessingEntry{}
	sb.log = log
	return sb
}

func (sb *simpleBuffer) SetFrame(frame BufferFrame) {
	sb.frame = frame
}

func (sb *simpleBuffer) Open() {
}

func (sb *simpleBuffer) Close() {
}

func (sb *simpleBuffer) Iterator() data.Iterator {
	return (&simpleBufferIterator{}).new(sb, true, sb.log)
}

func (sb *simpleBuffer) ProcessingIterator() ProcessingIterator {
	return (&simpleBufferIterator{}).new(sb, false, sb.log)
}

func (sb *simpleBuffer) Add(e ProcessingEntry) bool {
	sb.entries = append(sb.entries, e)
	return true
}

func (sb *simpleBuffer) Len() int {
	return len(sb.entries)
}

func (sb *simpleBuffer) Get(i int) interface{} {
	e := sb.entries[i]
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

func (sbi *simpleBufferIterator) new(buffer *simpleBuffer, valid bool, log logging.Context) *simpleBufferIterator {
	sbi.valid = valid
	sbi.current = -1
	sbi.buffer = buffer
	sbi.log = log
	return sbi
}

func (sbi *simpleBufferIterator) HasNext() bool {
	sbi.buffer.frame.Lock()
	defer sbi.buffer.frame.Unlock()
	for {
		sbi.log.Logger().Debug("HasNext", "current", sbi.current)
		if len(sbi.buffer.entries) > sbi.current+1 {
			if !sbi.valid || sbi.buffer.entries[sbi.current+1].Valid {
				return true
			}
			sbi.current++
			continue
		}
		if sbi.buffer.frame.IsClosed() {
			return false
		}
		sbi.buffer.frame.Wait()
	}
}

func (sbi *simpleBufferIterator) Next() interface{} {
	return sbi.NextProcessingEntry().Value
}

func (sbi *simpleBufferIterator) NextProcessingEntry() ProcessingEntry {
	sbi.buffer.frame.Lock()
	defer sbi.buffer.frame.Unlock()
	for {
		sbi.log.Logger().Debug("NextProcessingEntry", "current", sbi.current)
		if len(sbi.buffer.entries) > sbi.current+1 {
			sbi.current++
			if !sbi.valid || sbi.buffer.entries[sbi.current].Valid {
				return sbi.buffer.entries[sbi.current]
			}
			continue
		}
		if sbi.buffer.frame.IsClosed() {
			return ProcessingEntry{}
		}
		sbi.buffer.frame.Wait()
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

func (ob *orderedBuffer) new(log logging.Context) *orderedBuffer {
	(&ob.simple).new(log)
	ob.root.New(ob)
	ob.valid = ob.root.DLL()
	ob.last = ob.valid
	ob.nextIndex = ob.nextIndex.Next(-1, 0)
	ob.log = log
	return ob
}

func (ob *orderedBuffer) SetFrame(frame BufferFrame) {
	ob.simple.SetFrame(frame)
}

func (ob *orderedBuffer) Add(e ProcessingEntry) bool {
	e.Index.Validate(e.MaxIndex)
	ob.simple.Add(e)
	n := data.NewDLL(&e)

	c := ob.root.DLL()
	i := c.Next()
	for i != nil {
		v := i.Get().(*ProcessingEntry)
		if v.Index.After(e.Index) {
			break
		}
		c, i = i, i.Next()
	}
	c.Append(n)
	ob.size++
	if n.Next() == nil {
		ob.last = n
	}

	increased := false
	ob.log.Logger().Debug("add index to cur value", "index", e.Index, "value", e.Value, "next-index", ob.nextIndex)

	next := ob.valid.Next()
	for next != nil && !next.Get().(*ProcessingEntry).Index.After(ob.nextIndex) {
		n := next.Get().(*ProcessingEntry)
		ob.nextIndex = n.Index.Next(n.MaxIndex, n.MaxSub)
		ob.valid = next
		next = next.Next()
		increased = true
		ob.log.Logger().Debug("increase to index to value", "index", n.Index, "value", n.Value)
	}
	return increased
}

func (ob *orderedBuffer) Close() {
	ob.simple.Close()
	if ob.valid != ob.last {
		ob.valid = ob.last
		ob.nextIndex = ob.valid.Get().(*ProcessingEntry).Index
	}
}

func (ob *orderedBuffer) Open() {
	ob.simple.Open()
}

func (ob *orderedBuffer) Iterator() data.Iterator {
	return (&orderedBufferIterator{}).new(ob)
}

func (ob *orderedBuffer) ProcessingIterator() ProcessingIterator {
	return ob.simple.ProcessingIterator()
}

func (ob *orderedBuffer) Len() int {
	return ob.size
}

func (ob *orderedBuffer) Get(i int) interface{} {
	e := ob.root.DLL()
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

func (obi *orderedBufferIterator) new(buffer *orderedBuffer) *orderedBufferIterator {
	obi.buffer = buffer
	obi.current = obi.buffer.root.DLL()
	return obi
}

func (obi *orderedBufferIterator) HasNext() bool {
	obi.buffer.simple.frame.Lock()
	defer obi.buffer.simple.frame.Unlock()
	for {
		n := obi.current.Next()
		if n != nil && obi.current != obi.buffer.valid {
			if n.Get().(*ProcessingEntry).Valid {
				return true
			}
			obi.current = n // skip invalid entries
			continue
		}
		if obi.buffer.simple.frame.IsClosed() {
			return false
		}
		obi.buffer.simple.frame.Wait()
	}
}

func (obi *orderedBufferIterator) CheckNext() bool {
	obi.buffer.simple.frame.Lock()
	defer obi.buffer.simple.frame.Unlock()
	for {
		n := obi.current.Next()
		if n != nil && obi.current != obi.buffer.valid {
			if n.Get().(*ProcessingEntry).Valid {
				return true
			}
			obi.current = n // skip invalid entries
			continue
		}
		return false
	}
}

func (obi *orderedBufferIterator) Next() interface{} {
	obi.buffer.simple.frame.Lock()
	defer obi.buffer.simple.frame.Unlock()
	for {
		n := obi.current.Next()
		if n != nil && obi.current != obi.buffer.valid {
			e := n.Get().(*ProcessingEntry)
			obi.current = n // always proceed
			if e.Valid {
				return e.Value
			}
			continue
		}
		if obi.buffer.simple.frame.IsClosed() {
			return ProcessingEntry{}
		}
		obi.buffer.simple.frame.Wait()
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
