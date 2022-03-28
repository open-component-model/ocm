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
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
)

type IncrementalProcessingSource interface {
	data.Iterable
	Open()
	Add(e ...interface{}) IncrementalProcessingSource
	Close()
}

type ProcessingSource interface {
	IncrementalProcessingSource
	//data.IndexedAccess
}

type FilterFunction func(interface{}) bool
type MappingFunction data.MappingFunction
type ExplodeFunction func(interface{}) []interface{}
type CompareFunction = data.CompareFunction

func Identity(e interface{}) interface{} {
	return e
}

type ProcessingResult interface {
	data.Iterable

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

// Process processies an initial empty chain by converting
// an iterable into a ProcessingResult
func Process(data data.Iterable) ProcessingResult {
	return (&_SynchronousProcessing{}).new(data)
}

////////////////////////////////////////////////////////////////////////////

type _ProcessingSource struct {
	ProcessingBuffer
}

var _ ProcessingSource = &_ProcessingSource{}
var _ IncrementalProcessingSource = &_ProcessingSource{}
var _ data.IndexedAccess = &_ProcessingSource{}

func NewIncrementalProcessingSource() ProcessingSource {
	return (&_ProcessingSource{}).new()
}

func (this *_ProcessingSource) new() ProcessingSource {
	this.ProcessingBuffer = NewSimpleBuffer()
	return this
}

func (this *_ProcessingSource) Add(entries ...interface{}) IncrementalProcessingSource {
	for _, e := range entries {
		this.ProcessingBuffer.Add(NewEntry(Top(this.Len()), e))
	}
	return this
}

/////////////////////////////////////////////////////////////////////////////

func NewAsyncProcessingSource(f func() data.Iterable, pool ProcessorPool) ProcessingSource {
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
