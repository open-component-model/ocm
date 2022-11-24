// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package versionconstraintsoption

import (
	"github.com/Masterminds/semver/v3"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
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
	Latest      bool
	Constraints []*semver.Constraints
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Latest, "latest", "", false, "restrict component versions to latest")
	flag.SemverConstraintsVarP(fs, &o.Constraints, "constraints", "c", nil, "version constraint")
}

func (o *Option) Usage() string {

	s := `
If the option <code>--constraints</code> is given, and no version is specified for a component, only versions matching
the given version constraints (semver https://github.com/Masterminds/semver) are selected. With <code>--latest</code> only
the latest matching versions will be selected.
`
	return s
}
