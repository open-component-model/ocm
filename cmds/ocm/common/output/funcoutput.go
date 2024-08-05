package output

import (
	. "ocm.software/ocm/api/utils/out"
	. "ocm.software/ocm/cmds/ocm/common/processing"
)

type OutputFunction func(Context, interface{})

type FunctionProcessingOutput struct {
	ElementOutput
	function OutputFunction
}

var _ Output = &FunctionProcessingOutput{}

func NewProcessingFunctionOutput(opts *Options, chain ProcessChain, f OutputFunction) *FunctionProcessingOutput {
	return (&FunctionProcessingOutput{}).new(opts, chain, f)
}

func (this *FunctionProcessingOutput) new(opts *Options, chain ProcessChain, f OutputFunction) *FunctionProcessingOutput {
	this.ElementOutput.new(opts, chain)
	this.function = f
	return this
}

func (this *FunctionProcessingOutput) Out() error {
	i := this.Elems.Iterator()
	for i.HasNext() {
		this.function(this.Context, i.Next())
	}
	return this.ElementOutput.Out()
}
