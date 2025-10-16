package skipdigestoption

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm"
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
	flag *pflag.Flag
	Skip bool
}

var _ ocm.ModificationOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.flag = flag.BoolVarPF(fs, &o.Skip, "skip-digest-generation", "", false, "skip digest creation")
}

func (o *Option) Usage() string {
	s := `
If the option <code>--skip-digest-generation</code> is given, resources added to a
component version will not be digested, if no predefined digest is given. This
option should only be used to simulate legacy behaviour. Digests are required to
assure a proper transport behaviour.
`
	return s
}

func (o *Option) ApplyModificationOption(opts *ocm.ModificationOptions) {
	if o.flag == nil || o.flag.Changed {
		ocm.SkipDigest(o.Skip).ApplyModificationOption(opts) //nolint:staticcheck // skip digest still used for tests
	}
}
