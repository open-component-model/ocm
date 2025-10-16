package stoponexistingoption

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
	flag                  *pflag.Flag
	StopOnExistingVersion bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.flag = flag.BoolVarPF(fs, &o.StopOnExistingVersion, "stop-on-existing", "E", false, "stop on existing component version in target repository")
}

func (o *Option) Usage() string {
	s := `
If the option <code>--stop-on-existing</code> is given together with the <code>--recursive</code>
option, the recursion is stopped for component versions already existing in the 
target repository. This behaviour can be further influenced by specifying a transfer script
with the <code>script</code> option family.
`
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	if (o.flag != nil && o.flag.Changed) || o.StopOnExistingVersion {
		return standard.StopOnExistingVersion(o.StopOnExistingVersion).ApplyTransferOption(opts)
	}
	return nil
}
