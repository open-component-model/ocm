// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/processing"
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
