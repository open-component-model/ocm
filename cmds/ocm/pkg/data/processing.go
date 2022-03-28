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

type IncrementalProcessingSource interface {
	Iterable
	Open()
	Add(e ...interface{}) IncrementalProcessingSource
	Close()
}

type ProcessingSource interface {
	IncrementalProcessingSource
	IndexedAccess
}

type FilterFunction func(interface{}) bool
type MappingFunction func(interface{}) interface{}
type CompareFunction func(interface{}, interface{}) int

func Identity(e interface{}) interface{} {
	return e
}

type ProcessingResult interface {
	Iterable

	Map(m MappingFunction) ProcessingResult
	Filter(f FilterFunction) ProcessingResult
	Sort(c CompareFunction) ProcessingResult
	Apply(c ProcessChain) ProcessingResult

	Synchronously() ProcessingResult
	Asynchronously() ProcessingResult
	WithPool(ProcessorPool) ProcessingResult
	Unordered() ProcessingResult
	Parallel(n int) ProcessingResult

	AsSlice() IndexedSliceAccess
}


////////////////////////////////////////////////////////////////////////////

type _ProcessingSource struct {
	_container
}

var _ ProcessingSource = &_ProcessingSource{}
var _ IncrementalProcessingSource = &_ProcessingSource{}
var _ IndexedAccess = &_ProcessingSource{}

func NewIncrementalProcessingSource() ProcessingSource {
	return (&_ProcessingSource{}).new()
}

func (this *_ProcessingSource) new() ProcessingSource {
	this._container.new()
	return this
}

func (this *_ProcessingSource) Add(entries ...interface{}) IncrementalProcessingSource {
	for _, e := range entries {
		this.add(entry{this.len(), true, e})
	}
	return this
}

func (this *_ProcessingSource) Open()    { this.open() }
func (this *_ProcessingSource) Close()   { this.close() }
func (this *_ProcessingSource) Len() int { return this.len() }

func (this *_ProcessingSource) Get(i int) interface{} {
	e := this.get(i)
	return e.value
}

/////////////////////////////////////////////////////////////////////////////

func NewAsyncProcessingSource(f func() Iterable, pool ProcessorPool) ProcessingSource {
	p := (&_ProcessingSource{}).new()
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

