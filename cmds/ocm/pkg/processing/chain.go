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
	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
)

// ProcessChain is a data structure holding a chain definition, which is
// a chain of step creation functions used to instantiate the chain
// for a dedicated input processing.
// The instantiation is initiated by calling the Process
// method on a chain.
type ProcessChain interface {
	Transform(t TransformFunction) ProcessChain
	Explode(m ExplodeFunction) ProcessChain
	Map(m MappingFunction) ProcessChain
	Filter(f FilterFunction) ProcessChain
	Sort(c CompareFunction) ProcessChain
	WithPool(p ProcessorPool) ProcessChain
	Unordered() ProcessChain
	Parallel(n int) ProcessChain
	Append(p ProcessChain) ProcessChain

	Process(data data.Iterable) ProcessingResult
}

type stepCreator func(ProcessingResult) ProcessingResult

type _ProcessChain struct {
	parent  *_ProcessChain
	creator stepCreator
}

var _ ProcessChain = &_ProcessChain{}

func Chain() ProcessChain {
	return (&_ProcessChain{}).new(nil, nil)
}

func (this *_ProcessChain) new(p *_ProcessChain, creator stepCreator) *_ProcessChain {
	if p != nil {
		if p.creator != nil {
			this.parent = p
		} else {
			if p.parent != nil {
				this.parent = p.parent
			}
		}
	}
	if this.parent != nil && creator == nil {
		return this.parent
	}
	this.creator = creator
	return this
}

func (this *_ProcessChain) Transform(t TransformFunction) ProcessChain {
	if t == nil {
		return this
	}
	return (&_ProcessChain{}).new(this, chainTransform(t))
}

func (this *_ProcessChain) Explode(e ExplodeFunction) ProcessChain {
	if e == nil {
		return this
	}
	return (&_ProcessChain{}).new(this, chainExplode(e))
}
func (this *_ProcessChain) Map(m MappingFunction) ProcessChain {
	if m == nil {
		return this
	}
	return (&_ProcessChain{}).new(this, chainMap(m))
}
func (this *_ProcessChain) Filter(f FilterFunction) ProcessChain {
	if f == nil {
		return this
	}
	return (&_ProcessChain{}).new(this, chainFilter(f))
}
func (this *_ProcessChain) Sort(c CompareFunction) ProcessChain {
	if c == nil {
		return this
	}
	return (&_ProcessChain{}).new(this, chainSort(c))
}
func (this *_ProcessChain) WithPool(p ProcessorPool) ProcessChain {
	return (&_ProcessChain{}).new(this, chainWithPool(p))
}
func (this *_ProcessChain) Unordered() ProcessChain {
	return (&_ProcessChain{}).new(this, chainUnordered)
}
func (this *_ProcessChain) Parallel(n int) ProcessChain {
	return (&_ProcessChain{}).new(this, chainParallel(n))
}

func (this *_ProcessChain) Append(p ProcessChain) ProcessChain {
	if p == nil {
		return this
	}
	return (&_ProcessChain{}).new(this, chainApply(p))
}

// Process instantiates a processing chain for a dedicated input
// It builds a dedicated execution structure
// based on the chain functioned stored along the chain definition.
func (this *_ProcessChain) Process(data data.Iterable) ProcessingResult {
	p, ok := data.(ProcessingResult)
	if this.parent != nil {
		p = this.parent.Process(data)
	} else {
		if !ok {
			p = Process(data)
		}
	}
	return this.step(p)
}

func (this *_ProcessChain) step(p ProcessingResult) ProcessingResult {
	if this.creator == nil {
		return p
	}
	return this.creator(p)
}

func chainTransform(t TransformFunction) stepCreator {
	return func(p ProcessingResult) ProcessingResult { return p.Transform(t) }
}
func chainExplode(e ExplodeFunction) stepCreator {
	return func(p ProcessingResult) ProcessingResult { return p.Explode(e) }
}
func chainMap(m MappingFunction) stepCreator {
	return func(p ProcessingResult) ProcessingResult { return p.Map(m) }
}
func chainFilter(f FilterFunction) stepCreator {
	return func(p ProcessingResult) ProcessingResult { return p.Filter(f) }
}
func chainSort(c CompareFunction) stepCreator {
	return func(p ProcessingResult) ProcessingResult { return p.Sort(c) }
}
func chainWithPool(pool ProcessorPool) stepCreator {
	return func(p ProcessingResult) ProcessingResult { return p.WithPool(pool) }
}
func chainParallel(n int) stepCreator {
	return func(p ProcessingResult) ProcessingResult { return p.Parallel(n) }
}
func chainUnordered(p ProcessingResult) ProcessingResult { return p.Unordered() }

func chainApply(c ProcessChain) stepCreator {
	return func(p ProcessingResult) ProcessingResult {
		return p.Apply(c)
	}
}

////////////////////////////////////////////////////////////////////////////////

var initial = Chain()

func Transform(t TransformFunction) ProcessChain { return initial.Transform(t) }
func Explode(e ExplodeFunction) ProcessChain     { return initial.Explode(e) }
func Map(m MappingFunction) ProcessChain         { return initial.Map(m) }
func Filter(f FilterFunction) ProcessChain       { return initial.Filter(f) }
func Sort(c CompareFunction) ProcessChain        { return initial.Sort(c) }
func WithPool(pool ProcessorPool) ProcessChain   { return initial.WithPool(pool) }
func Parallel(n int) ProcessChain                { return initial.Parallel(n) }
func Unordered() ProcessChain                    { return initial.Unordered() }
func Append(chain, add ProcessChain, conditions ...options.Condition) ProcessChain {
	if add != nil {
		if options.And(conditions...).IsTrue() {
			if chain == nil {
				return add
			}
			return chain.Append(add)
		}
	}
	return chain
}
