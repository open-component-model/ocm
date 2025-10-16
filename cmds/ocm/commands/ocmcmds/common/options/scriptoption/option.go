package scriptoption

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/spiff"
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
	spiff.TransferOptionsCreator
	ScriptFile string
	Script     string
	ScriptData []byte
	FileSystem vfs.FileSystem
}

var (
	_ options.OptionWithCLIContextCompleter = (*Option)(nil)
	_ transferhandler.TransferOption        = (*Option)(nil)
)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ScriptFile, "scriptFile", "s", "", "filename of transfer handler script")
	fs.StringVarP(&o.ScriptFile, "script", "", "", "config name of transfer handler script")
}

func (o *Option) Configure(ctx clictx.Context) error {
	o.FileSystem = ctx.FileSystem()
	if o.ScriptFile != "" && o.Script != "" {
		return errors.Newf("only one of --script or --scriptFile may be set")
	}
	if o.ScriptData != nil {
		return nil
	}
	if o.Script != "" {
		err := cfgcpi.NewUpdater(ctx.ConfigContext(), o).Update()
		if err != nil {
			return err
		}
		if o.ScriptData == nil {
			return errors.ErrUnknown("script", o.Script)
		}
	}
	if o.ScriptFile != "" {
		data, err := utils.ReadFile(o.ScriptFile, ctx.FileSystem())
		if err != nil {
			return errors.Wrapf(err, "invalid transfer script file")
		}
		o.ScriptData = data
	}
	if o.ScriptData == nil {
		o.Script = "default"
		err := cfgcpi.NewUpdater(ctx.ConfigContext(), o).Update()
		if o.ScriptData == nil {
			o.Script = ""
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Option) Usage() string {
	s := `
It is possible to use a dedicated transfer script based on spiff.
The option <code>--scriptFile</code> can be used to specify this script
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

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	var err error
	if o.ScriptData != nil {
		err = spiff.Script(o.ScriptData).ApplyTransferOption(opts)
	} else if o.ScriptFile != "" {
		err = spiff.ScriptByFile(o.ScriptFile, o.FileSystem).ApplyTransferOption(opts)
	}
	return err
}
