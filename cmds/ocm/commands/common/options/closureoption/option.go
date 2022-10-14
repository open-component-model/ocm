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
	"fmt"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

type Option struct {
	ElementName      string
	Closure          bool
	ClosureField     string
	AddReferencePath options.OptionSelector
	AdditionalFields []string
	FieldEnricher    func(interface{}) []string
}

func New(elemname string, settings ...interface{}) *Option {
	o := &Option{ElementName: elemname, AddReferencePath: options.Always()}
	for _, s := range settings {
		switch v := s.(type) {
		case options.OptionSelector:
			o.AddReferencePath = v
		case string:
			o.ClosureField = v
		case []string:
			o.AdditionalFields = v
		case func(interface{}) []string:
			o.FieldEnricher = v
		default:
			panic(fmt.Errorf("invalid setting for closure option: %T", s))
		}
	}
	if (len(o.AdditionalFields) > 0) != (o.FieldEnricher != nil) {
		panic(fmt.Errorf("invalid setting for closure option: both, addituonal fields and enricher must be set"))
	}
	return o
}

func (o *Option) IsTrue() bool {
	return o.Closure
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	return standard.Recursive(o.Closure).ApplyTransferOption(opts)
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Closure, "recursive", "r", false, fmt.Sprintf("follow %s nesting", o.ElementName))
}

func (o *Option) Usage() string {
	return fmt.Sprintf("\nWith the option <code>--recursive</code> the complete reference tree of a %s is traversed.\n", o.ElementName)
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

func (o *Option) Headers(opts options.OptionSetProvider, cols []string) []string {
	if o.Closure {
		if o.AddReferencePath(opts) {
			h := o.ClosureField
			if h == "" {
				h = "REFERENCEPATH"
			}
			cols = insert(cols, h)
		}
		return append(cols, o.AdditionalFields...)
	}
	return cols
}

func (o *Option) additionalFields(e interface{}) []string {
	if o.FieldEnricher != nil {
		return o.FieldEnricher(e)
	}
	return nil
}

func (o *Option) Mapper(opts options.OptionSetProvider, path func(interface{}) string, mapper processing.MappingFunction) processing.MappingFunction {
	if o.Closure {
		use := o.AddReferencePath(opts)
		return func(e interface{}) interface{} {
			fields := mapper(e).([]string)
			if use {
				fields = insert(fields, path(e))
			}
			return append(fields, o.additionalFields(e)...)
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
	return processing.Append(chain, Chain(opts, cf))
}

func Chain(opts *output.Options, cf ClosureFunction) processing.ProcessChain {
	return processing.Explode(cf.Exploder(opts))
}

// OutputChainFunction provides an chain function that can be used to add an option
// based closure processing and an optional additional chain to an output chain.
func OutputChainFunction(cf ClosureFunction, chain processing.ProcessChain) output.ChainFunction {
	return func(opts *output.Options) processing.ProcessChain {
		return Closure(opts, cf, chain)
	}
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

func AddChain(opts *output.Options, chain, add processing.ProcessChain) processing.ProcessChain {
	copts := From(opts)

	if !copts.Closure {
		return chain
	}
	return processing.Append(chain, add)
}

func TableOutput(in *output.TableOutput, closure ...ClosureFunction) *output.TableOutput {
	var cf ClosureFunction
	if len(closure) > 0 {
		cf = closure[0]
	}
	chain := processing.Append(in.Chain, processing.Explode(cf.Exploder(in.Options)))
	copts := From(in.Options)
	return &output.TableOutput{
		Headers: copts.Headers(in.Options, in.Headers),
		Options: in.Options,
		Chain:   chain,
		Mapping: copts.Mapper(in.Options, History, in.Mapping),
	}
}
