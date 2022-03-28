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

// ProcessChain is a data structure holding a chain definition, which is
// a chain of step creation functions used to instantiate the chain
// for a dedicated input processing.
// The instantiation is initiated by calling the Process
// method on a chain.
type ProcessChain interface {
	Explode(m ExplodeFunction) ProcessChain
	Map(m MappingFunction) ProcessChain
	Filter(f FilterFunction) ProcessChain
	Sort(c CompareFunction) ProcessChain
	WithPool(p ProcessorPool) ProcessChain
	Unordered() ProcessChain
	Parallel(n int) ProcessChain

	Process(data data.Iterable) ProcessingResult
}

type step_creator func(ProcessingResult) ProcessingResult

type _ProcessChain struct {
	parent  *_ProcessChain
	creator step_creator
}

var _ ProcessChain = &_ProcessChain{}

func Chain() ProcessChain {
	return (&_ProcessChain{}).new(nil, nil)
}

func (this *_ProcessChain) new(p *_ProcessChain, creator step_creator) *_ProcessChain {
	this.parent = p
	this.creator = creator
	return this
}

func (this *_ProcessChain) Explode(e ExplodeFunction) ProcessChain {
	return (&_ProcessChain{}).new(this, chain_explode(e))
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

// Process instantiates a processing chain for a dedicated input
// It builds a dedicated execution structure
// based on the chain functioned stored along the chain definition.
func (this *_ProcessChain) Process(data data.Iterable) ProcessingResult {
	p, ok := data.(ProcessingResult)
	if ok {
		if this.parent == nil {
			return p
		}
		return this.creator(this.parent.Process(p))
	}
	if this.parent == nil {
		return Process(data)
	}
	return this.creator(this.parent.Process(data))
}

func chain_explode(e ExplodeFunction) step_creator {
	return func(p ProcessingResult) ProcessingResult { return p.Explode(e) }
}
func chain_map(m MappingFunction) step_creator {
	return func(p ProcessingResult) ProcessingResult { return p.Map(m) }
}
func chain_filter(f FilterFunction) step_creator {
	return func(p ProcessingResult) ProcessingResult { return p.Filter(f) }
}
func chain_sort(c CompareFunction) step_creator {
	return func(p ProcessingResult) ProcessingResult { return p.Sort(c) }
}
func chain_with_pool(pool ProcessorPool) step_creator {
	return func(p ProcessingResult) ProcessingResult { return p.WithPool(pool) }
}
func chain_parallel(n int) step_creator {
	return func(p ProcessingResult) ProcessingResult { return p.Parallel(n) }
}
func chain_unordered(p ProcessingResult) ProcessingResult { return p.Unordered() }

func IdentityMapper(e interface{}) interface{} {
	return e
}
