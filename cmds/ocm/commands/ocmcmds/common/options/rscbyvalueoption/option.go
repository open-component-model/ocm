// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rscbyvalueoption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
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
	ResourcesByValue bool
	LocalByValue     bool
}

var _ transferhandler.TransferOption = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.ResourcesByValue, "copy-resources", "V", false, "transfer referenced resources by-value")
	fs.BoolVarP(&o.LocalByValue, "copy-local-resources", "L", false, "transfer referenced local resources by-value")
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
	err := standard.ResourcesByValue(o.ResourcesByValue).ApplyTransferOption(opts)
	if err == nil {
		err = standard.LocalResourcesByValue(o.LocalByValue).ApplyTransferOption(opts)
	}
	return err
}
