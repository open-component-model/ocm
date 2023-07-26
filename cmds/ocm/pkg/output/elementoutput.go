// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	. "github.com/open-component-model/ocm/v2/cmds/ocm/pkg/processing"
	. "github.com/open-component-model/ocm/v2/pkg/out"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/data"
)

type ElementOutput struct {
	source  ProcessingSource
	Elems   data.Iterable
	Context Context
}

var _ Output = (*ElementOutput)(nil)

func NewElementOutput(log logging.Context, ctx Context, chain ProcessChain) *ElementOutput {
	return (&ElementOutput{}).new(log, ctx, chain)
}

func (this *ElementOutput) new(log logging.Context, ctx Context, chain ProcessChain) *ElementOutput {
	if log == nil {
		log = logging.DefaultContext()
	}
	this.source = NewIncrementalProcessingSource(log)
	this.Context = ctx
	if chain == nil {
		this.Elems = this.source
	} else {
		this.Elems = Process(log, this.source).Asynchronously().Apply(chain)
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

func (this *ElementOutput) Out() error {
	return nil
}
