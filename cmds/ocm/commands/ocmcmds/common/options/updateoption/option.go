// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package updateoption

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

type Option struct {
	flag   *pflag.Flag
	Update bool
}

func New() *Option {
	return &Option{}
}

func (o *Option) IsTrue() bool {
	return o.Update
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	if o.flag != nil && o.flag.Changed {
		return standard.Update(!o.Update).ApplyTransferOption(opts)
	}
	return nil
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.flag = flag.BoolVarPF(fs, &o.Update, "no-update", "", false, "don't touch existing versions in target")
}

func (o *Option) Usage() string {
	return fmt.Sprintf(`
With the option <code>--no-update</code> existing versions in the target
repository will not be touched at all. An additional specification of the
option <code>--overwrite</code> is ignored. By default, updates of
volative (non-signature-relevant) information is enabled, but the
modification of non-volatile data is prohibited unless the overwrite
option is given.
`)
}
