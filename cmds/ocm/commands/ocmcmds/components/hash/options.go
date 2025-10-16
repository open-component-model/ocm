package hash

import (
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/tools/signing"
	common "ocm.software/ocm/api/utils/misc"
	signingcmd "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/cmds/signing"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/hashoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

type Option struct {
	Actual bool
	Update bool
	Verify bool

	action  signingcmd.Action
	outfile string
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Actual, "actual", "", false, "use actual component descriptor")
	fs.BoolVarP(&o.Update, "update", "U", false, "update digests in component version")
	fs.BoolVarP(&o.Verify, "verify", "V", false, "verify digests found in component version")
	fs.StringVarP(&o.outfile, "outfile", "O", "-", "Output file for normalized component descriptor")
}

func (o *Option) Complete(cmd *Command) error {
	if o.Actual {
		return nil
	}
	repo := repooption.From(cmd).Repository
	lookup := lookupoption.From(cmd)
	sopts := signing.NewOptions(hashoption.From(cmd), signing.Resolver(repo, lookup.Resolver), signing.Update(o.Update), signing.VerifyDigests(o.Verify))
	err := sopts.Complete(cmd.Context.OCMContext())
	if err == nil {
		o.action = signingcmd.NewAction([]string{"", ""}, cmd.Context.OCMContext(), common.NewPrinter(nil), sopts)
	}
	return err
}

func (o *Option) Usage() string {
	s := `
If the option <code>--actual</code> is given the component descriptor actually
found is used as it is, otherwise the required digests are calculated on-the-fly.
`
	return s
}
