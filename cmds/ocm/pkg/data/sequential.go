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
	"sync"
)

func Process(data Iterable) ProcessingResult {
	return (&_SynchronousProcessing{}).new(data)
}

////////////////////////////////////////////////////////////////////////////

type operation interface {
	process(e interface{}) (interface{}, bool)
}

type mapper MappingFunction

func (this mapper) process(e interface{}) (interface{}, bool) {
	return this(e), true
}

type filter FilterFunction

func (this filter) process(e interface{}) (interface{}, bool) {
	if this(e) {
		return e, true
	}
	return nil, false
}

/////////////////////////////////////////////////////////////////////////////

type _SynchronousProcessing struct {
	data Iterable
}

var _ Iterable = &_SynchronousProcessing{}

func (this *_SynchronousProcessing) new(data Iterable) *_SynchronousProcessing {
	this.data = data
	return this
}

func (this *_SynchronousProcessing) Map(m MappingFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this, process(mapper(m)))
}
func (this *_SynchronousProcessing) Filter(f FilterFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this, process(filter(f)))
}
func (this *_SynchronousProcessing) Sort(c CompareFunction) ProcessingResult {
	return (&_SynchronousStep{}).new(this, process_sort(c))
}
func (this *_SynchronousProcessing) WithPool(p ProcessorPool) ProcessingResult {
	return (&_ParallelProcessing{}).new(newEntryIterableFromIterable(this.data), p, NewOrderedContainer)
}
func (this *_SynchronousProcessing) Parallel(n int) ProcessingResult {
	return this.WithPool(NewProcessorPool(n))
}
func (this *_SynchronousProcessing) Synchronously() ProcessingResult {
	return this
}
func (this *_SynchronousProcessing) Asynchronously() ProcessingResult {
	return (&_AsynchronousProcessing{}).new(this)
}
func (this *_SynchronousProcessing) Unordered() ProcessingResult {
	return this
}
func (this *_SynchronousProcessing) Apply(c ProcessChain) ProcessingResult {
	return c.Process(this)
}

func (this *_SynchronousProcessing) Iterator() Iterator {
	return this.data.Iterator()
}
func (this *_SynchronousProcessing) AsSlice() IndexedSliceAccess {
	return IndexedSliceAccess(Slice(this.data))
}

////////////////////////////////////////////////////////////////////////////

type _SynchronousStep struct {
	_SynchronousProcessing
}

func (this *_SynchronousStep) new(data Iterable, proc processing) *_SynchronousStep {
	this.data = proc(data)
	return this
}

/////////////////////////////////////////////////////////////////////////////

type processing func(Iterable) Iterable

type _AsynchronousProcessing struct {
	data Iterable
	lock sync.Mutex
}

var _ Iterable = &_AsynchronousProcessing{}

func (this *_AsynchronousProcessing) new(data Iterable) *_AsynchronousProcessing {
	this.data = data
	return this
}

func (this *_AsynchronousProcessing) Map(m MappingFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, process(mapper(m)))
}
func (this *_AsynchronousProcessing) Filter(f FilterFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, process(filter(f)))
}
func (this *_AsynchronousProcessing) Sort(c CompareFunction) ProcessingResult {
	return (&_AsynchronousStep{}).new(this, process_sort(c))
}
func (this *_AsynchronousProcessing) WithPool(p ProcessorPool) ProcessingResult {
	return (&_ParallelProcessing{}).new(newEntryIterableFromIterable(this.data), p, NewOrderedContainer)
}
func (this *_AsynchronousProcessing) Parallel(n int) ProcessingResult {
	return this.WithPool(NewProcessorPool(n))
}
func (this *_AsynchronousProcessing) Synchronously() ProcessingResult {
	return (&_SynchronousProcessing{}).new(this)
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

func (this *_AsynchronousProcessing) Iterator() Iterator {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.data.Iterator()
}
func (this *_AsynchronousProcessing) AsSlice() IndexedSliceAccess {
	return IndexedSliceAccess(Slice(this))
}

////////////////////////////////////////////////////////////////////////////

type _AsynchronousStep struct {
	_AsynchronousProcessing
}

func (this *_AsynchronousStep) new(data Iterable, proc processing) *_AsynchronousStep {
	this.lock.Lock()
	go func() {
		this.data = proc(data)
		this.lock.Unlock()
	}()

	return this
}

////////////////////////////////////////////////////////////////////////////

func process_sort(c CompareFunction) func(data Iterable) Iterable {
	return func(data Iterable) Iterable {
		slice := Slice(data)
		Sort(slice, c)
		return IndexedSliceAccess(slice)
	}
}

func process(op operation) processing {
	return func(data Iterable) Iterable {
		slice := []interface{}{}
		i := data.Iterator()
		for i.HasNext() {
			e, ok := op.process(i.Next())
			if ok {
				slice = append(slice, e)
			}
		}
		return IndexedSliceAccess(slice)
	}
}
