// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"github.com/spf13/pflag"

	signingcmd "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/cmds/signing"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/hashoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

type Option struct {
	Actual bool

	action  signingcmd.Action
	outfile string
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Actual, "actual", "", false, "use actual component descriptor")
	fs.StringVarP(&o.outfile, "outfile", "O", "norm.ncd", "Output file for normalized component descriptor")
}

func (o *Option) Complete(cmd *Command) error {
	if o.Actual {
		return nil
	}
	repo := repooption.From(cmd).Repository
	lookup := lookupoption.From(cmd)
	sopts := signing.NewOptions(hashoption.From(cmd), signing.Resolver(repo, lookup.Resolver))
	err := sopts.Complete(signingattr.Get(cmd.Context.OCMContext()))
	if err == nil {
		o.action = signingcmd.NewAction([]string{"", ""}, common.NewPrinter(nil), sopts)
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
