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
	. "github.com/gardener/ocm/cmds/ocm/pkg/data"
	. "github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	"github.com/gardener/ocm/pkg/errors"
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
	this.ElementOutput.new(opts.Context, chain)
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
	return nil
}
