// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package templateroption

import (
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(def string) *Option {
	return &Option{template.Options{Default: def}}
}

type Option struct {
	template.Options
}

func (o *Option) Configure(ctx clictx.Context) error {
	return o.Options.Complete(ctx.FileSystem())
}
