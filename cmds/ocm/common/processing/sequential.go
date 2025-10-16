package processing

import (
	"sync"

	"github.com/mandelsoft/logging"
	"ocm.software/ocm/cmds/ocm/common/data"
)

type _SynchronousProcessing struct {
	data data.Iterable
	log  logging.Context
}

var _ data.Iterable = &_SynchronousProcessing{}

func (this *_SynchronousProcessing) new(log logging.Context, data data.Iterable) *_SynchronousProcessing {
	this.data = data
	this.log = log
	return this
}

func (this *_SynchronousProcessing) Transform(t TransformFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this.log, this, t)
}

func (this *_SynchronousProcessing) Explode(e ExplodeFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this.log, this, process(explode(e)))
}

func (this *_SynchronousProcessing) Map(m MappingFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this.log, this, process(mapper(m)))
}

func (this *_SynchronousProcessing) Filter(f FilterFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this.log, this, process(filter(f)))
}

func (this *_SynchronousProcessing) Sort(c CompareFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this.log, this, processSort(c))
}

func (this *_SynchronousProcessing) WithPool(p ProcessorPool) ProcessingResult {
	return (&_ParallelProcessing{}).new(NewEntryIterableFromIterable(this.log, this.data), p, NewOrderedBuffer, this.log)
}

func (this *_SynchronousProcessing) Parallel(n int) ProcessingResult {
	return this.WithPool(NewProcessorPool(n))
}

func (this *_SynchronousProcessing) Synchronously() ProcessingResult {
	return this
}

func (this *_SynchronousProcessing) Asynchronously() ProcessingResult {
	return (&_AsynchronousProcessing{}).new(this.log, this)
}

func (this *_SynchronousProcessing) Unordered() ProcessingResult {
	return this
}

func (this *_SynchronousProcessing) Apply(c ProcessChain) ProcessingResult {
	return c.Process(this)
}

func (this *_SynchronousProcessing) Iterator() data.Iterator {
	return this.data.Iterator()
}

func (this *_SynchronousProcessing) AsSlice() data.IndexedSliceAccess {
	return data.Slice(this.data)
}

type _SynchronousStep struct {
	_SynchronousProcessing
}

func (this *_SynchronousStep) new(log logging.Context, data data.Iterable, proc processing) *_SynchronousStep {
	this.data = proc(data)
	this.log = log
	return this
}

type processing func(data.Iterable) data.Iterable

type _AsynchronousProcessing struct {
	data data.Iterable
	lock sync.Mutex
	log  logging.Context
}

var _ data.Iterable = &_AsynchronousProcessing{}

func (this *_AsynchronousProcessing) new(log logging.Context, data data.Iterable) *_AsynchronousProcessing {
	this.data = data
	this.log = log
	return this
}

func (this *_AsynchronousProcessing) Transform(t TransformFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, t)
}

func (this *_AsynchronousProcessing) Explode(m ExplodeFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, process(explode(m)))
}

func (this *_AsynchronousProcessing) Map(m MappingFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, process(mapper(m)))
}

func (this *_AsynchronousProcessing) Filter(f FilterFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, process(filter(f)))
}

func (this *_AsynchronousProcessing) Sort(c CompareFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, processSort(c))
}

func (this *_AsynchronousProcessing) WithPool(p ProcessorPool) ProcessingResult {
	return (&_ParallelProcessing{}).new(NewEntryIterableFromIterable(this.log, this.data), p, NewOrderedBuffer, this.log)
}

func (this *_AsynchronousProcessing) Parallel(n int) ProcessingResult {
	return this.WithPool(NewProcessorPool(n))
}

func (this *_AsynchronousProcessing) Synchronously() ProcessingResult {
	return (&_SynchronousProcessing{}).new(this.log, this)
}

func (this *_AsynchronousProcessing) Asynchronously() ProcessingResult {
	return this
}

func (this *_AsynchronousProcessing) Unordered() ProcessingResult {
	return this
}

func (this *_AsynchronousProcessing) Apply(c ProcessChain) ProcessingResult {
	return c.Process(this)
}

func (this *_AsynchronousProcessing) Iterator() data.Iterator {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.data.Iterator()
}

func (this *_AsynchronousProcessing) AsSlice() data.IndexedSliceAccess {
	return data.IndexedSliceAccess(data.Slice(this))
}

type _AsynchronousStep struct {
	_AsynchronousProcessing
}

func (this *_AsynchronousStep) new(data data.Iterable, proc processing) *_AsynchronousStep {
	this.lock.Lock()
	go func() {
		this.data = proc(data)
		this.lock.Unlock()
	}()

	return this
}

func processSort(c CompareFunction) func(data data.Iterable) data.Iterable {
	return func(it data.Iterable) data.Iterable {
		slice := data.Slice(it)
		data.Sort(slice, c)
		return data.IndexedSliceAccess(slice)
	}
}

func process(op operation) processing {
	return func(it data.Iterable) data.Iterable {
		slice := []interface{}{}
		i := it.Iterator()
		for i.HasNext() {
			e, ok := op.process(i.Next())
			if ok {
				switch len(e) {
				case 0:
					slice = append(slice, nil)
				default:
					slice = append(slice, e...)
				}
			}
		}
		return data.IndexedSliceAccess(slice)
	}
}
