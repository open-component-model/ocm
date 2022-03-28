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

package output

import (
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	. "github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	. "github.com/gardener/ocm/cmds/ocm/pkg/processing"
)

type ElementOutput struct {
	source  ProcessingSource
	Elems   data.Iterable
	Context Context
}

func NewElementOutput(ctx Context, chain ProcessChain) *ElementOutput {
	return (&ElementOutput{}).new(ctx, chain)
}

func (this *ElementOutput) new(ctx Context, chain ProcessChain) *ElementOutput {
	this.source = NewIncrementalProcessingSource()
	this.Context = ctx
	if chain == nil {
		this.Elems = this.source
	} else {
		this.Elems = Process(this.source).Asynchronously().Apply(chain)
	}
	return this
}

func (this *ElementOutput) Add(e interface{}) error {
	this.source.Add(e)
	return nil
}

func (this *ElementOutput) Close() error {
	this.source.Close()
	return nil
}

func (this *ElementOutput) Out() {
}
