package destoption

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
	Destination    string
	PathFilesystem vfs.FileSystem
}

func (d *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&d.Destination, "outfile", "O", "", "output file or directory")
}

func (o *Option) Configure(ctx clictx.Context) error {
	o.PathFilesystem = ctx.FileSystem()
	return nil
}

var _ options.Options = (*Option)(nil)
