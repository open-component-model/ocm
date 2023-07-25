// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rscbyvalueoption

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
	rflag            *pflag.Flag
	lflag            *pflag.Flag
	ResourcesByValue bool
	LocalByValue     bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.rflag = flag.BoolVarPF(fs, &o.ResourcesByValue, "copy-resources", "V", false, "transfer referenced resources by-value")
	o.lflag = flag.BoolVarPF(fs, &o.LocalByValue, "copy-local-resources", "L", false, "transfer referenced local resources by-value")
}

func (o *Option) Usage() string {
	s := `
It the option <code>--copy-resources</code> is given, all referential 
resources will potentially be localized, mapped to component version local
resources in the target repository. It the option <code>--copy-local-resources</code> 
is given, instead, only resources with the relation <code>local</code> will be
transferred. This behaviour can be further influenced by specifying a transfer
script with the <code>script</code> option family.
`
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	var err error
	if (o.rflag != nil && o.rflag.Changed) || o.ResourcesByValue {
		err = standard.ResourcesByValue(o.ResourcesByValue).ApplyTransferOption(opts)
	}
	if err == nil {
		if (o.lflag != nil && o.lflag.Changed) || o.LocalByValue {
			err = standard.LocalResourcesByValue(o.LocalByValue).ApplyTransferOption(opts)
		}
	}
	return err
}
