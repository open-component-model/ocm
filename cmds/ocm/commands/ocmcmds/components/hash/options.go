// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

type Option struct {
	Actual bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Actual, "actual", "", false, "use actual component descriptor")
}

func (o *Option) Usage() string {
	s := `
If the option <code>--actual</code> is given the component descriptor actually
found is used as it is, otherwise the required digests are calculated on-the-fly.
`
	return s
}
