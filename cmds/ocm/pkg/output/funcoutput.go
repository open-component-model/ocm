// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"github.com/mandelsoft/logging"
	. "github.com/open-component-model/ocm/v2/cmds/ocm/pkg/processing"
)

type OutputFunction func(Context, interface{})

type FunctionProcessingOutput struct {
	ElementOutput
	function OutputFunction
}

var _ Output = &FunctionProcessingOutput{}

func NewProcessingFunctionOutput(log logging.Context, ctx Context, chain ProcessChain, f OutputFunction) *FunctionProcessingOutput {
	return (&FunctionProcessingOutput{}).new(log, ctx, chain, f)
}

func (this *FunctionProcessingOutput) new(log logging.Context, ctx Context, chain ProcessChain, f OutputFunction) *FunctionProcessingOutput {
	this.ElementOutput.new(log, ctx, chain)
	this.function = f
	return this
}

func (this *FunctionProcessingOutput) Out() error {
	i := this.Elems.Iterator()
	for i.HasNext() {
		this.function(this.Context, i.Next())
	}
	return nil
}
