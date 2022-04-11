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

package closureoption

import (
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/processing"
	"github.com/gardener/ocm/pkg/common"
	"github.com/spf13/pflag"
)

func From(o *output.Options) *Option {
	var opt *Option
	o.Get(&opt)
	return opt
}

type Option struct {
	Closure          bool
	ClosureField     string
	AdditionalFields []string
	FieldEnricher    func(interface{}) []string
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Closure, "closure", "c", false, "follow component references")
}

func (o *Option) Usage() string {
	return `
With the option <code>--closure</code> the complete reference tree by a component verserion is traversed.
`
}

func (o *Option) Explode(e processing.ExplodeFunction) processing.ProcessChain {
	if o.Closure {
		return processing.Explode(e)
	}
	return nil
}

func insert(a []string, v string) []string {
	r := make([]string, len(a)+1)
	copy(r[1:], a)
	r[0] = v
	return r
}

func (o *Option) Headers(cols []string) []string {
	h := o.ClosureField
	if h == "" {
		h = "REFERENCEPATH"
	}
	if o.Closure {
		return append(insert(cols, h), o.AdditionalFields...)
	}
	return cols
}

func (o *Option) additionalFields(e interface{}) []string {
	if o.FieldEnricher != nil {
		return o.FieldEnricher(e)
	}
	return nil
}

func (o *Option) Mapper(path func(interface{}) string, mapper processing.MappingFunction) processing.MappingFunction {
	if o.Closure {
		return func(e interface{}) interface{} {
			return append(insert(mapper(e).([]string), path(e)), o.additionalFields(e)...)
		}
	}
	return mapper
}

func History(e interface{}) string {
	if o, ok := e.(common.HistorySource); ok {
		if h := o.GetHistory(); h != nil {
			return h.String()
		}
	}
	return ""
}

func Closure(opts *output.Options, cf ClosureFunction, chain processing.ProcessChain) processing.ProcessChain {
	return processing.Append(chain, processing.Explode(cf.Exploder(opts)))
}

type ClosureFunction func(*output.Options, interface{}) []interface{}

func (c ClosureFunction) Exploder(opts *output.Options) processing.ExplodeFunction {
	if c != nil {
		copts := From(opts)
		if copts.Closure {
			return func(e interface{}) []interface{} { return c(opts, e) }
		}
	}
	return nil
}

func TableOutput(in *output.TableOutput, closure ...ClosureFunction) *output.TableOutput {
	var cf ClosureFunction
	if len(closure) > 0 {
		cf = closure[0]
	}
	chain := processing.Append(in.Chain, processing.Explode(cf.Exploder(in.Options)))
	copts := From(in.Options)
	return &output.TableOutput{
		Headers: copts.Headers(in.Headers),
		Options: in.Options,
		Chain:   chain,
		Mapping: copts.Mapper(History, in.Mapping),
	}
}
