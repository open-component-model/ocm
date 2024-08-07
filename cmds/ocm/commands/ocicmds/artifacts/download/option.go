package download

import (
	"github.com/spf13/pflag"

	"ocm.software/ocm/cmds/ocm/common/options"
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
	Layers  []int
	DirTree bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.IntSliceVarP(&o.Layers, "layers", "", nil, "extract dedicated layers")
	fs.BoolVarP(&o.DirTree, "dirtree", "", false, "extract as effective filesystem content")
}

func (o *Option) Usage() string {
	return `
With option <code>--layers</code> it is possible to request the download of
dedicated layers, only. Option <code>--dirtree</code> expects the artifact to
be a layered filesystem (for example OCI Image) and provided the effective
filesystem content.
`
}
