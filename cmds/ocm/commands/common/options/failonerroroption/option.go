// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package failonerroroption

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/spf13/pflag"
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
	Fail bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Fail, "fail-on-error", "", false, "fail on label validation error")
}

var _ options.Options = (*Option)(nil)
