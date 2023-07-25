// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package overwriteoption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/v2/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/transfer/transferhandler/standard"
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
	flag      *pflag.Flag
	Overwrite bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.flag = flag.BoolVarPF(fs, &o.Overwrite, "overwrite", "f", false, "overwrite existing component versions")
}

func (o *Option) Usage() string {
	s := `
It the option <code>--overwrite</code> is given, component version in the
target repository will be overwritten, if they already exist.
`
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	if (o.flag != nil && o.flag.Changed) || o.Overwrite {
		return standard.Overwrite(o.Overwrite).ApplyTransferOption(opts)
	}
	return nil
}
