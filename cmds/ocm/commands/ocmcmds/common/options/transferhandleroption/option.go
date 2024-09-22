package transferhandleroption

import (
	"strings"

	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
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

var _ options.OptionWithCLIContextCompleter = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.setting, "transferHandler", "T", "", "transfer handler (<name>[=<config>)")
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
It is possible to use a dedicated transfer handlers,
either buil-in one or plugin based handlers.
The option <code>--transferHandler</code> can be used to specify this script
by a file name. With <code>--script</code> it can be taken from the 
CLI config using an entry of the following format:

<pre>
type: scripts.ocm.config.ocm.software
scripts:
  &lt;name>: 
    path: &lt;filepath> 
    script:
      &lt;scriptdata>
</pre>

Only one of the fields <code>path</code> or <code>script</code> can be used.

If no script option is given and the cli config defines a script <code>default</code>
this one is used.
`
	return s
}
