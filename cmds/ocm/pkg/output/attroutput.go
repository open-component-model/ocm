package output

import (
	. "github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/pkg/errors"
	. "github.com/open-component-model/ocm/pkg/out"
)

type AttrProcessingOutput struct {
	ElementOutput
	mapper func(interface{}) *AttributeSet
	opts   *Options
}

var _ Output = &AttrProcessingOutput{}

func NewProcessingAttrOutput(opts *Options, chain ProcessChain, header ...string) *AttrProcessingOutput {
	return (&AttrProcessingOutput{}).new(opts, chain, header)
}

func (this *AttrProcessingOutput) new(opts *Options, chain ProcessChain, header []string) *AttrProcessingOutput {
	this.ElementOutput.new(opts, chain)
	this.opts = opts
	return this
}

func (this *AttrProcessingOutput) Out() error {
	var ok bool
	i := this.Elems.Iterator()
	for i.HasNext() {
		Outf(this.opts.Context, "---\n")
		elem := i.Next()
		var set *AttributeSet
		if this.mapper != nil {
			set = this.mapper(elem)
		} else {
			set, ok = i.Next().(*AttributeSet)
			if !ok {
				return errors.Newf("invalid attr type")
			}
		}
		set.PrintAttributes(this.opts.Context)
	}
	return this.ElementOutput.Out()
}
