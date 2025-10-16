package get

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func AttachedFrom(o options.OptionSetProvider) *Attached {
	var opt *Attached
	o.AsOptionSet().Get(&opt)
	return opt
}

type Attached struct {
	Flag bool
}

var (
	_ options.Condition = (*Attached)(nil)
	_ options.Options   = (*Attached)(nil)
)

func (a *Attached) IsTrue() bool {
	return a.Flag
}

func (a *Attached) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&a.Flag, "attached", "a", false, "show attached artifacts")
}
