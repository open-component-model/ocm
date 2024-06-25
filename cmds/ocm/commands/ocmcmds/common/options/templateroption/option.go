package templateroption

import (
	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/api/utils/template"
	"github.com/open-component-model/ocm/cmds/ocm/common/options"
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
