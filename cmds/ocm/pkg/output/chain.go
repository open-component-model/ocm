package output

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
)

type ChainFunction func(opts *Options) processing.ProcessChain

func ComposeChain(funcs ...ChainFunction) ChainFunction {
	return func(opts *Options) processing.ProcessChain {
		var chain processing.ProcessChain
		for _, f := range funcs {
			chain = processing.Append(chain, f(opts))
		}
		return chain
	}
}
