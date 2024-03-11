// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package check

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/check"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

type Option struct {
	CheckLocalResources bool
	CheckLocalSources   bool
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) ApplyTo(opts *check.Options) {
	optionutils.ApplyOption(&o.CheckLocalSources, &opts.CheckLocalSources)
	optionutils.ApplyOption(&o.CheckLocalResources, &opts.CheckLocalResources)
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.CheckLocalResources, "local-resources", "R", false, "check also for describing resources with local access method, only")
	fs.BoolVarP(&o.CheckLocalSources, "local-sources", "S", false, "check also for describing sources with local access method, only")
}

func (o *Option) Usage() string {
	s := `
If the options <code>--local-resources</code> and/or <code>--local-sources</code> are given the
the check additionally assures that all resources or sources are included into the component version.
This means that they are using local access methods, only.
`
	return s
}
