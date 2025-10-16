package keepglobaloption

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
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
	KeepGlobalAccess bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.KeepGlobalAccess, "keep-global-access", "G", false, "preserve global access for value transport")
}

func (o *Option) Usage() string {
	s := `
If the option <code>--keep-global-access</code> is given, all localized referential 
resources will preserve their original global access information.
This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.
`
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	return standard.KeepGlobalAccess(o.KeepGlobalAccess).ApplyTransferOption(opts)
}
