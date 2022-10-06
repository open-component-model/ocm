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
	"sync"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

type _ParallelProcessing struct {
	data    ProcessingIterable
	pool    ProcessorPool
	creator BufferCreator
	ctx     ocm.Context
}

var _ data.Iterable = &_ParallelProcessing{}

func (this *_ParallelProcessing) new(data ProcessingIterable, pool ProcessorPool, creator BufferCreator, ctx ocm.Context) *_ParallelProcessing {
	this.data = data
	this.pool = pool
	this.creator = creator
	this.ctx = ctx
	return this
}

func (this *_ParallelProcessing) Explode(m ExplodeFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.ctx, this.pool, this.data, explode(m), this.creator)
}

func (this *_ParallelProcessing) Map(m MappingFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.ctx, this.pool, this.data, mapper(m), this.creator)
}

func (this *_ParallelProcessing) Filter(f FilterFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.ctx, this.pool, this.data, filter(f), this.creator)
}

func (this *_ParallelProcessing) Sort(c CompareFunction) ProcessingResult {
	setup := func() data.Iterable { return this.AsSlice().Sort(c) }

	this.ctx.Logger().Debug("sorting pool", "pool", this.pool)

	return (&_ParallelProcessing{}).new(NewAsyncProcessingSource(this.ctx, setup, this.pool).(ProcessingIterable), this.pool, NewOrderedBuffer, this.ctx)
}

func (this *_ParallelProcessing) Transform(t TransformFunction) ProcessingResult {
	transform := func() data.Iterable { return t(this.data) }
	return (&_ParallelProcessing{}).new(NewAsyncProcessingSource(this.ctx, transform, this.pool).(ProcessingIterable), this.pool, NewOrderedBuffer, this.ctx)
}

func (this *_ParallelProcessing) WithPool(p ProcessorPool) ProcessingResult {
	return (&_ParallelProcessing{}).new(this.data, p, this.creator, this.ctx)
}

func (this *_ParallelProcessing) Parallel(n int) ProcessingResult {
	return this.WithPool(NewProcessorPool(n))
}

func (this *_ParallelProcessing) Synchronously() ProcessingResult {
	return (&_SynchronousProcessing{}).new(this.ctx, this)
}

func (this *_ParallelProcessing) Asynchronously() ProcessingResult {
	return (&_AsynchronousProcessing{}).new(this.ctx, this)
}

func (this *_ParallelProcessing) Unordered() ProcessingResult {
	data := this.data
	ordered, ok := data.(*orderedBuffer)
	if ok {
		data = &ordered.simple
	}
	return (&_ParallelProcessing{}).new(data, this.pool, NewSimpleBuffer, this.ctx)
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

func (this *_ParallelStep) new(ctx ocm.Context, pool ProcessorPool, data ProcessingIterable, op operation, creator BufferCreator) *_ParallelStep {
	this.ctx = ctx
	buffer := creator(ctx)
	this._ParallelProcessing.new(buffer, pool, creator, this.ctx)
	go func() {
		this.ctx.Logger().Debug("start processing")

		this.pool.Request()
		i := data.ProcessingIterator()
		var wg sync.WaitGroup
		for i.HasNext() {
			e := i.NextProcessingEntry()
			this.ctx.Logger().Debug("start", "index", e.Index)
			wg.Add(1)
			pool.Exec(func() {
				this.ctx.Logger().Debug("process", "index", e.Index)
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
				this.ctx.Logger().Debug("done", "index", e.Index)
				wg.Done()
			})
		}
		wg.Wait()
		this.pool.Release()
		buffer.Close()
		this.ctx.Logger().Debug("done processing")
	}()
	return this
}
