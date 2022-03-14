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

type ProcessChain interface {
	Map(m MappingFunction) ProcessChain
	Filter(f FilterFunction) ProcessChain
	Sort(c CompareFunction) ProcessChain
	WithPool(p ProcessorPool) ProcessChain
	Unordered() ProcessChain
	Parallel(n int) ProcessChain

	Process(data Iterable) ProcessingResult
}

type chain_operation func(ProcessingResult) ProcessingResult

type _ProcessChain struct {
	parent    *_ProcessChain
	operation chain_operation
}

var _ ProcessChain = &_ProcessChain{}

func Chain() ProcessChain {
	return (&_ProcessChain{}).new(nil, nil)
}

func (this *_ProcessChain) new(p *_ProcessChain, op chain_operation) *_ProcessChain {
	this.parent = p
	this.operation = op
	return this
}

func (this *_ProcessChain) Map(m MappingFunction) ProcessChain {
	return (&_ProcessChain{}).new(this, chain_map(m))
}
func (this *_ProcessChain) Filter(f FilterFunction) ProcessChain {
	return (&_ProcessChain{}).new(this, chain_filter(f))
}
func (this *_ProcessChain) Sort(c CompareFunction) ProcessChain {
	return (&_ProcessChain{}).new(this, chain_sort(c))
}
func (this *_ProcessChain) WithPool(p ProcessorPool) ProcessChain {
	return (&_ProcessChain{}).new(this, chain_with_pool(p))
}
func (this *_ProcessChain) Unordered() ProcessChain {
	return (&_ProcessChain{}).new(this, chain_unordered)
}
func (this *_ProcessChain) Parallel(n int) ProcessChain {
	return (&_ProcessChain{}).new(this, chain_parallel(n))
}

func (this *_ProcessChain) Process(data Iterable) ProcessingResult {
	p, ok := data.(ProcessingResult)
	if ok {
		if this.parent == nil {
			return p
		}
		return this.operation(this.parent.Process(p))
	}
	if this.parent == nil {
		return Process(data)
	}
	return this.operation(this.parent.Process(data))
}

func chain_map(m MappingFunction) chain_operation {
	return func(p ProcessingResult) ProcessingResult { return p.Map(m) }
}
func chain_filter(f FilterFunction) chain_operation {
	return func(p ProcessingResult) ProcessingResult { return p.Filter(f) }
}
func chain_sort(c CompareFunction) chain_operation {
	return func(p ProcessingResult) ProcessingResult { return p.Sort(c) }
}
func chain_with_pool(pool ProcessorPool) chain_operation {
	return func(p ProcessingResult) ProcessingResult { return p.WithPool(pool) }
}
func chain_parallel(n int) chain_operation {
	return func(p ProcessingResult) ProcessingResult { return p.Parallel(n) }
}
func chain_unordered(p ProcessingResult) ProcessingResult { return p.Unordered() }
