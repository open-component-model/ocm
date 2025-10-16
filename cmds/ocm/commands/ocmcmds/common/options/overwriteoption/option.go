package overwriteoption

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

	overwrite *pflag.Flag
	Overwrite bool

	enforce          *pflag.Flag
	EnforceTransport bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.overwrite = flag.BoolVarPF(fs, &o.Overwrite, "overwrite", "f", false, "overwrite existing component versions")
	o.enforce = flag.BoolVarPF(fs, &o.EnforceTransport, "enforce", "", false, "enforce transport as if target version were not present")
}

func (o *Option) Usage() string {
	s := `
If the option <code>--overwrite</code> is given, component versions in the
target repository will be overwritten, if they already exist, but with different digest.
If the option <code>--enforce</code> is given, component versions in the
target repository will be transported as if they were not present on the target side,
regardless of their state (this is independent on their actual state, even identical 
versions are re-transported).
`
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	if (o.overwrite != nil && o.overwrite.Changed) || o.Overwrite {
		if err := standard.Overwrite(o.Overwrite).ApplyTransferOption(opts); err != nil {
			return err
		}
	}
	if (o.enforce != nil && o.enforce.Changed) || o.EnforceTransport {
		if err := standard.EnforceTransport(o.EnforceTransport).ApplyTransferOption(opts); err != nil {
			return err
		}
	}
	return nil
}
