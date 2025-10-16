package rscbyvalueoption

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
	rflag            *pflag.Flag
	lflag            *pflag.Flag
	ResourcesByValue bool
	LocalByValue     bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.rflag = flag.BoolVarPF(fs, &o.ResourcesByValue, "copy-resources", "V", false, "transfer referenced resources by-value")
	o.lflag = flag.BoolVarPF(fs, &o.LocalByValue, "copy-local-resources", "L", false, "transfer referenced local resources by-value")
}

func (o *Option) Usage() string {
	s := `
If the option <code>--copy-resources</code> is given, all referential 
resources will potentially be localized, mapped to component version local
resources in the target repository. If the option <code>--copy-local-resources</code> 
is given, instead, only resources with the relation <code>local</code> will be
transferred. This behaviour can be further influenced by specifying a transfer
script with the <code>script</code> option family.
`
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	var err error
	if (o.rflag != nil && o.rflag.Changed) || o.ResourcesByValue {
		err = standard.ResourcesByValue(o.ResourcesByValue).ApplyTransferOption(opts)
	}
	if err == nil {
		if (o.lflag != nil && o.lflag.Changed) || o.LocalByValue {
			err = standard.LocalResourcesByValue(o.LocalByValue).ApplyTransferOption(opts)
		}
	}
	return err
}
