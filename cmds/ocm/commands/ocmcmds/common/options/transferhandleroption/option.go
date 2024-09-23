package transferhandleroption

import (
	"strings"

	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/utils"
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
	setting string
	Path    string
	Config  []byte
}

var (
	_ options.OptionWithCLIContextCompleter = (*Option)(nil)
	_ transferhandler.TransferOption        = (*Option)(nil)
)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.setting, "transfer-handler", "T", "", "transfer handler (<name>[=<config>)")
}

func (o *Option) Configure(ctx clictx.Context) error {
	if o.setting == "" {
		return nil
	}
	idx := strings.Index(o.setting, "=")
	if idx >= 0 {
		o.Path = o.setting[:idx]
		data, err := utils.ResolveData(o.setting[idx+1:], vfsattr.Get(ctx))
		if err != nil {
			return err
		}
		o.Config = data
	} else {
		o.Path = o.setting
	}
	return nil
}

func (o *Option) Usage() string {
	s := `
It is possible to use dedicated transfer handlers, either built-in ones or
plugin based handlers. The option <code>--transferHandler</code> can be used to specify
this handler using the hierarchical handler notation scheme.
` + transferhandler.Usage(ocm.DefaultContext())
	return s
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	var err error
	if len(o.Config) != 0 {
		err = transferhandler.WithConfig(o.Config).ApplyTransferOption(opts)
	}
	return err
}
