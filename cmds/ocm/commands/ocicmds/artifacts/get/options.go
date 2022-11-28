// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
)

func AttachedFrom(o options.OptionSetProvider) *Attached {
	var opt *Attached
	o.AsOptionSet().Get(&opt)
	return opt
}

type Attached struct {
	Flag bool
}

var (
	_ options.Condition = (*Attached)(nil)
	_ options.Options   = (*Attached)(nil)
)

func (a *Attached) IsTrue() bool {
	return a.Flag
}

func (a *Attached) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&a.Flag, "attached", "a", false, "show attached artifacts")
}
