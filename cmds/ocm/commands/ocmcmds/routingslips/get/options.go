package get

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func NewOptions() *Option {
	return &Option{}
}

type Option struct {
	Verify bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Verify, "verify", "v", false, "verify signature")
}

var _ options.Options = (*Option)(nil)
