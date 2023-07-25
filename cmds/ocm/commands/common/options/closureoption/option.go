// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package closureoption

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/v2/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/v2/pkg/common"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/v2/pkg/utils"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

type Option struct {
	flag *pflag.Flag

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
	if (o.flag != nil && o.flag.Changed) || o.Closure {
		return standard.Recursive(o.Closure).ApplyTransferOption(opts)
	}
	return nil
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.flag = flag.BoolVarPF(fs, &o.Closure, "recursive", "r", false, fmt.Sprintf("follow %s nesting", o.ElementName))
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
	cf := utils.Optional(closure...)
	chain := processing.Append(in.Chain, processing.Explode(cf.Exploder(in.Options)))
	copts := From(in.Options)
	return &output.TableOutput{
		Headers: copts.Headers(in.Options, in.Headers),
		Options: in.Options,
		Chain:   chain,
		Mapping: copts.Mapper(in.Options, History, in.Mapping),
	}
}
