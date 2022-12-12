// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dryrunoption

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

type Option struct {
	out     bool
	usage   string
	Outfile string
	DryRun  bool
}

func New(usage string, out bool) *Option {
	if usage == "" {
		usage = "dry-run mode"
	}
	return &Option{
		out:   out,
		usage: usage,
	}
}
func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.DryRun, "dry-run", "", false, o.usage)
	if o.out {
		fs.StringVarP(&o.Outfile, "output", "O", "", "output file for dry-run")
	}
}

func (o *Option) Complete() error {
	if o.Outfile != "" && !o.DryRun {
		return fmt.Errorf("--output only usable for dry-run mode")
	}
	return nil
}
