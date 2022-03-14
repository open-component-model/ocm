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
	"fmt"
	"sync"
)

var log = false

type _ParallelProcessing struct {
	data    entry_iterable
	pool    ProcessorPool
	creator container_creator
}

var _ Iterable = &_ParallelProcessing{}

func (this *_ParallelProcessing) new(data entry_iterable, pool ProcessorPool, creator container_creator) *_ParallelProcessing {
	this.data = data
	this.pool = pool
	this.creator = creator
	return this
}

func (this *_ParallelProcessing) Map(m MappingFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.pool, this.data, mapper(m), this.creator)
}
func (this *_ParallelProcessing) Filter(f FilterFunction) ProcessingResult {
	return (&_ParallelStep{}).new(this.pool, this.data, filter(f), this.creator)
}
func (this *_ParallelProcessing) Sort(c CompareFunction) ProcessingResult {
	setup := func() Iterable { return this.AsSlice().Sort(c) }
	fmt.Printf("POOL %+v\n", this.pool)
	return (&_ParallelProcessing{}).new(NewAsyncProcessingSource(setup, this.pool).(entry_iterable), this.pool, NewOrderedContainer)
}

func (this *_ParallelProcessing) WithPool(p ProcessorPool) ProcessingResult {
	return (&_ParallelProcessing{}).new(this.data, p, this.creator)
}
func (this *_ParallelProcessing) Parallel(n int) ProcessingResult {
	return this.WithPool(NewProcessorPool(n))
}
func (this *_ParallelProcessing) Synchronously() ProcessingResult {
	return (&_SynchronousProcessing{}).new(this)
}
func (this *_ParallelProcessing) Asynchronously() ProcessingResult {
	return (&_AsynchronousProcessing{}).new(this)
}
func (this *_ParallelProcessing) Unordered() ProcessingResult {
	data := this.data
	ordered, ok := data.(*ordered_container)
	if ok {
		data = &ordered._container
	}
	return (&_ParallelProcessing{}).new(data, this.pool, NewContainer)
}
func (this *_ParallelProcessing) Apply(p ProcessChain) ProcessingResult {
	return p.Process(this)
}

func (this *_ParallelProcessing) Iterator() Iterator {
	return this.data.Iterator()
}
func (this *_ParallelProcessing) AsSlice() IndexedSliceAccess {
	return IndexedSliceAccess(Slice(this.data))
}

////////////////////////////////////////////////////////////////////////////

type _ParallelStep struct {
	_ParallelProcessing
	container container
	op        operation
	create    container_creator
}

func (this *_ParallelStep) new(pool ProcessorPool, data entry_iterable, op operation, creator container_creator) *_ParallelStep {
	this.container = creator()
	this._ParallelProcessing.new(this.container, pool, creator)
	go func() {
		if log {
			fmt.Printf("start processing\n")
		}
		this.pool.Request()
		i := data.entry_iterator()
		var wg sync.WaitGroup
		for i.HasNext() {
			e := i.next()
			if log {
				fmt.Printf("start %d\n", e.index)
			}
			wg.Add(1)
			pool.Exec(func() {
				if log {
					fmt.Printf("process %d\n", e.index)
				}
				e.value, e.ok = op.process(e.value)
				this.container.add(e)
				if log {
					fmt.Printf("done %d\n", e.index)
				}
				wg.Done()

			})
		}
		wg.Wait()
		this.pool.Release()
		this.container.close()
		if log {
			fmt.Printf("done processing\n")
		}
	}()
	return this
}
