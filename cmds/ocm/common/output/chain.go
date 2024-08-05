package output

import (
	"ocm.software/ocm/cmds/ocm/common/processing"
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
