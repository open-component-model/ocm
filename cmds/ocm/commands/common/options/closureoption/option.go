package closureoption

import (
	"fmt"

	"github.com/modern-go/reflect2"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	if !reflect2.IsNil(o) {
		o.AsOptionSet().Get(&opt)
	}
	return opt
}

type Option struct {
	standard.TransferOptionsCreator
	flag *pflag.Flag

	ElementName      string
	Closure          bool
	ClosureField     string
	AddReferencePath options.OptionSelector
	AdditionalFields []string
	FieldEnricher    func(interface{}) []string
}

var _ transferhandler.TransferOption = (*Option)(nil)

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
	if o != nil && o.Closure {
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
		if copts != nil && copts.Closure {
			return func(e interface{}) []interface{} { return c(opts, e) }
		}
	}
	return nil
}

func AddChain(opts *output.Options, chain, add processing.ProcessChain) processing.ProcessChain {
	copts := From(opts)

	if copts == nil || !copts.Closure {
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
