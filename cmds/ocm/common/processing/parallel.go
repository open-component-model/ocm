package processing

import (
	"sync"

	"github.com/mandelsoft/logging"

	"ocm.software/ocm/cmds/ocm/common/data"
)

type _ParallelProcessing struct {
	data    ProcessingIterable
	pool    ProcessorPool
	creator BufferCreator
	log     logging.Context
}

var _ data.Iterable = &_ParallelProcessing{}

func (this *_ParallelProcessing) new(data ProcessingIterable, pool ProcessorPool, creator BufferCreator, log logging.Context) *_ParallelProcessing {
	this.data = data
	this.pool = pool
	this.creator = creator
	this.log = log
	return this
}

func (this *_ParallelProcessing) Explode(m ExplodeFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.log, this.pool, this.data, explode(m), this.creator)
}

func (this *_ParallelProcessing) Map(m MappingFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.log, this.pool, this.data, mapper(m), this.creator)
}

func (this *_ParallelProcessing) Filter(f FilterFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.log, this.pool, this.data, filter(f), this.creator)
}

func (this *_ParallelProcessing) Sort(c CompareFunction) ProcessingResult {
	setup := func() data.Iterable { return this.AsSlice().Sort(c) }

	this.log.Logger().Debug("sorting pool", "pool", this.pool)

	return (&_ParallelProcessing{}).new(NewAsyncProcessingSource(this.log, setup, this.pool).(ProcessingIterable), this.pool, NewOrderedBuffer, this.log)
}

func (this *_ParallelProcessing) Transform(t TransformFunction) ProcessingResult {
	transform := func() data.Iterable { return t(this.data) }
	return (&_ParallelProcessing{}).new(NewAsyncProcessingSource(this.log, transform, this.pool).(ProcessingIterable), this.pool, NewOrderedBuffer, this.log)
}

func (this *_ParallelProcessing) WithPool(p ProcessorPool) ProcessingResult {
	return (&_ParallelProcessing{}).new(this.data, p, this.creator, this.log)
}

func (this *_ParallelProcessing) Parallel(n int) ProcessingResult {
	return this.WithPool(NewProcessorPool(n))
}

func (this *_ParallelProcessing) Synchronously() ProcessingResult {
	return (&_SynchronousProcessing{}).new(this.log, this)
}

func (this *_ParallelProcessing) Asynchronously() ProcessingResult {
	return (&_AsynchronousProcessing{}).new(this.log, this)
}

func (this *_ParallelProcessing) Unordered() ProcessingResult {
	data := this.data
	ordered, ok := data.(*orderedBuffer)
	if ok {
		data = &ordered.simple
	}
	return (&_ParallelProcessing{}).new(data, this.pool, NewSimpleBuffer, this.log)
}

func (this *_ParallelProcessing) Apply(p ProcessChain) ProcessingResult {
	return p.Process(this)
}

func (this *_ParallelProcessing) Iterator() data.Iterator {
	return this.data.Iterator()
}

func (this *_ParallelProcessing) AsSlice() data.IndexedSliceAccess {
	return data.IndexedSliceAccess(data.Slice(this.data))
}

////////////////////////////////////////////////////////////////////////////

type _ParallelStep struct {
	_ParallelProcessing
}

func (this *_ParallelStep) new(log logging.Context, pool ProcessorPool, data ProcessingIterable, op operation, creator BufferCreator) *_ParallelStep {
	this.log = log
	buffer := creator(log)
	this._ParallelProcessing.new(buffer, pool, creator, this.log)
	go func() {
		this.log.Logger().Debug("start processing")

		this.pool.Request()
		i := data.ProcessingIterator()
		var wg sync.WaitGroup
		for i.HasNext() {
			e := i.NextProcessingEntry()
			this.log.Logger().Debug("start", "index", e.Index)
			wg.Add(1)
			pool.Exec(func() {
				this.log.Logger().Debug("process", "index", e.Index)
				var r operationResult
				if e.Valid {
					r, e.Valid = op.process(e.Value)
				}
				if !e.Valid {
					e.Value = nil
					// keep indicating index with for unused value
					buffer.Add(e)
				} else {
					switch len(r) {
					case 0:
						e.Value = nil
						buffer.Add(e)
					case 1:
						e.Value = r[0]
						buffer.Add(e)
					default:
						sub := len(r)
						// first: indicate number of sub entries with unused dummy entry
						buffer.Add(NewEntry(e.Index, nil, SubEntries(sub), e.MaxIndex, false))
						// second: apply all sub entries
						for idx, n := range r {
							buffer.Add(NewEntry(append(e.Index.Copy(), idx), n, sub))
						}
					}
				}
				this.log.Logger().Debug("done", "index", e.Index)
				wg.Done()
			})
		}
		wg.Wait()
		this.pool.Release()
		buffer.Close()
		this.log.Logger().Debug("done processing")
	}()
	return this
}
