package plugin

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
	return o.plugin.Command(o.name, o.StdOut(), o.args)
}
