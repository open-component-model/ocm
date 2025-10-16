package processing

import (
	"github.com/mandelsoft/logging"
	"ocm.software/ocm/cmds/ocm/common/data"
)

type IncrementalProcessingSource interface {
	data.Iterable
	Open()
	Add(e ...interface{}) IncrementalProcessingSource
	Close()
}

type ProcessingSource interface {
	IncrementalProcessingSource
	// data.IndexedAccess
}

type (
	FilterFunction         func(interface{}) bool
	MappingFunction        data.MappingFunction
	ExplodeFunction        func(interface{}) []interface{}
	CompareFunction        = data.CompareFunction
	TransformFunction      = func(iterable data.Iterable) data.Iterable
	CompareIndexedFunction = data.CompareIndexedFunction
	AggregationFunction    func(e, aggr interface{}) interface{}
)

func Identity(e interface{}) interface{} {
	return e
}

type ProcessingResult interface {
	data.Iterable

	Transform(t TransformFunction) ProcessingResult
	Explode(e ExplodeFunction) ProcessingResult
	Map(m MappingFunction) ProcessingResult
	Filter(f FilterFunction) ProcessingResult
	Sort(c CompareFunction) ProcessingResult
	Apply(c ProcessChain) ProcessingResult

	Synchronously() ProcessingResult
	Asynchronously() ProcessingResult
	WithPool(ProcessorPool) ProcessingResult
	Unordered() ProcessingResult
	Parallel(n int) ProcessingResult

	AsSlice() data.IndexedSliceAccess
}

////////////////////////////////////////////////////////////////////////////

// Process processes an initial empty chain by converting
// an iterable into a ProcessingResult.
func Process(log logging.Context, data data.Iterable) ProcessingResult {
	return (&_SynchronousProcessing{}).new(log, data)
}

////////////////////////////////////////////////////////////////////////////

type _ProcessingSource struct {
	ProcessingBuffer
}

var (
	_ ProcessingSource            = &_ProcessingSource{}
	_ IncrementalProcessingSource = &_ProcessingSource{}
	_ data.IndexedAccess          = &_ProcessingSource{}
)

func NewIncrementalProcessingSource(log logging.Context) ProcessingSource {
	return (&_ProcessingSource{}).new(log)
}

func (this *_ProcessingSource) new(log logging.Context) ProcessingSource {
	this.ProcessingBuffer = NewSimpleBuffer(log)
	return this
}

func (this *_ProcessingSource) Add(entries ...interface{}) IncrementalProcessingSource {
	for _, e := range entries {
		this.ProcessingBuffer.Add(NewEntry(Top(this.Len()), e))
	}
	return this
}

/////////////////////////////////////////////////////////////////////////////

func NewAsyncProcessingSource(log logging.Context, f func() data.Iterable, pool ProcessorPool) ProcessingSource {
	p := (&_ProcessingSource{}).new(log)
	pool.Request()
	pool.Exec(func() {
		i := f().Iterator()
		for i.HasNext() {
			p.Add(i.Next())
		}
		p.Close()
		pool.Release()
	})
	return p
}

func MappingSequence(mapper ...MappingFunction) MappingFunction {
	return func(e interface{}) interface{} {
		for _, m := range mapper {
			e = m(e)
		}
		return e
	}
}
