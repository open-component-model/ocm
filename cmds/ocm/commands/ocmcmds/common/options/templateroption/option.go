package templateroption

import (
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils/template"
	"ocm.software/ocm/cmds/ocm/common/options"
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
