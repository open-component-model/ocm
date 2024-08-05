package plugin

import (
	"strings"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

type Command struct {
	utils.BaseCommand
	plugin plugin.Plugin
	pcmd   *plugin.CommandDescriptor
	name   string

	args []string
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new plugin based command.
func NewCommand(ctx clictx.Context, plugin plugin.Plugin, name string, names ...string) *cobra.Command {
	me := &Command{BaseCommand: utils.NewBaseCommand(ctx)}
	me.plugin = plugin
	me.name = name
	me.pcmd = plugin.GetDescriptor().Commands.Get(name)

	cmd := utils.SetupCommand(me, utils.Names([]string{me.pcmd.Name}, names...)...)
	cmd.DisableFlagParsing = true
	cmd.SetHelpFunc(me.help)
	return cmd
}

func (o *Command) ForName(name string) *cobra.Command {
	pcmd := o.plugin.GetDescriptor().Commands.Get(o.name)
	return &cobra.Command{
		Use:     pcmd.Usage[strings.Index(pcmd.Usage, " ")+1:],
		Short:   pcmd.Short,
		Long:    pcmd.GetDescription(),
		Example: pcmd.Example,
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
}

func (o *Command) Complete(args []string) error {
	o.args = args
	return nil
}

func (o *Command) Run() error {
	return o.plugin.Command(o.name, o.StdIn(), o.StdOut(), o.args)
}

func (o *Command) help(cmd *cobra.Command, args []string) {
	o.plugin.Command(o.name, o.StdIn(), o.StdOut(), sliceutils.CopyAppend(args, "--help"))
}
