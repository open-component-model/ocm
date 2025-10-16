package srcbyvalueoption

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New() *Option {
	return &Option{}
}

type Option struct {
	standard.TransferOptionsCreator
	flag           *pflag.Flag
	SourcesByValue bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.flag = flag.BoolVarPF(fs, &o.SourcesByValue, "copy-sources", "", false, "transfer referenced sources by-value")
}

func (o *Option) Usage() string {
	s := `
If the option <code>--copy-sources</code> is given, all referential 
sources will potentially be localized, mapped to component version local
resources in the target repository.
This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.
`
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	if (o.flag != nil && o.flag.Changed) || o.SourcesByValue {
		return standard.SourcesByValue(o.SourcesByValue).ApplyTransferOption(opts)
	}
	return nil
}
