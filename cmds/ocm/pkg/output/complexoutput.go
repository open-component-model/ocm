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
	"fmt"

	. "github.com/gardener/ocm/cmds/ocm/pkg/data"
	. "github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	"github.com/gardener/ocm/pkg/utils"

	"github.com/gardener/ocm/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type ComplexProcessingOutput struct {
	ElementOutput
	mapper func(interface{}) interface{}
	fields []string
}

var _ Output = &ComplexProcessingOutput{}

func NewProcessingComplexOutput(ctx Context, chain ProcessChain, fields ...string) *ComplexProcessingOutput {
	return (&ComplexProcessingOutput{}).new(ctx, chain, fields)
}

func (this *ComplexProcessingOutput) new(ctx Context, chain ProcessChain, fields []string) *ComplexProcessingOutput {
	this.ElementOutput.new(ctx, chain)
	this.fields = fields
	return this
}

func (this *ComplexProcessingOutput) Out() error {
	i := this.Elems.Iterator()
	for i.HasNext() {
		Outf(this.Context, "---\n")
		elem := i.Next()
		var out interface{}
		if this.mapper != nil {
			out = this.mapper(elem)

		}
		data, err := runtime.DefaultYAMLEncoding.Marshal(out)
		if err != nil {
			Error(this.Context, err.Error())
		} else {
			if len(this.fields) > 0 {
				m := map[string]interface{}{}
				runtime.DefaultYAMLEncoding.Unmarshal(data, m)
				this.out("", m)

			} else {
				Outf(this.Context, "%s\n", string(data))
			}
		}
	}
	return nil
}

func (this *ComplexProcessingOutput) out(gap string, m map[string]interface{}) {
	rest := map[string]interface{}{}
	for k, v := range m {
		rest[k] = v
	}

	for _, k := range this.fields {
		v := m[k]
		delete(rest, k)
		if v != nil {
			switch e := v.(type) {
			case map[string]interface{}:
				Outf(this.Context, "%s%s:\n", gap, k)
				this.out(gap+"  ", e)
			case []interface{}:
				Outf(this.Context, "%s%s:\n", gap, k)
				s, err := yaml.Marshal(v)
				if err == nil {
					utils.IndentLines(string(s), gap)
				}
			default:
				eff := utils.IndentLines(fmt.Sprintf("%v", v), gap+"  ")
				Outf(this.Context, "%s%s: %s", gap, k, eff[len(gap)+2:])
			}
		}
	}
	s, err := yaml.Marshal(rest)
	if err == nil {
		Outf(this.Context, utils.IndentLines(string(s), gap))
	}
}
