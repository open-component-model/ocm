//go:generate go run -mod=mod ./doc ../../../../../docs/pluginreference

package cmds

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/action"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/command"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/describe"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/download"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/info"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/mergehandler"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/topics/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/upload"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/valueset"
	"ocm.software/ocm/api/utils/cobrautils"
)

type PluginCommand struct {
	command *cobra.Command
	plugin  ppi.Plugin
}

func (p *PluginCommand) Command() *cobra.Command {
	return p.command
}

func NewPluginCommand(p ppi.Plugin) *PluginCommand {
	short := p.Descriptor().Short
	if short == "" {
		short = "OCM plugin " + p.Name()
	}

	pcmd := &PluginCommand{
		plugin: p,
	}
	cmd := &cobra.Command{
		Use:                   p.Name() + " <subcommand> <options> <args>",
		Short:                 short,
		Long:                  p.Descriptor().Long,
		Version:               p.Version(),
		PersistentPreRunE:     pcmd.PreRunE,
		TraverseChildren:      true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		SilenceErrors:         true,
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cobrautils.TweakCommand(cmd, nil)

	cmd.AddCommand(describe.New(p))
	cmd.AddCommand(info.New(p))
	cmd.AddCommand(action.New(p))
	cmd.AddCommand(mergehandler.New(p))
	cmd.AddCommand(accessmethod.New(p))
	cmd.AddCommand(upload.New(p))
	cmd.AddCommand(download.New(p))
	cmd.AddCommand(valueset.New(p))
	cmd.AddCommand(command.New(p))

	cmd.InitDefaultHelpCmd()
	help := cobrautils.GetHelpCommand(cmd)

	// help.Use="help <topic>"
	help.DisableFlagsInUseLine = true
	cmd.AddCommand(descriptor.New())

	help.AddCommand(descriptor.New())

	p.GetOptions().AddFlags(cmd.Flags())
	pcmd.command = cmd
	return pcmd
}

type Error struct {
	Error string `json:"error"`
}

func (p *PluginCommand) PreRunE(cmd *cobra.Command, args []string) error {
	if handler != nil {
		return handler.HandleConfig(p.plugin.GetOptions().LogConfig)
	}
	return nil
}

func (p *PluginCommand) Execute(args []string) error {
	p.command.SetArgs(args)
	err := p.command.Execute()
	if err != nil {
		result, err2 := json.Marshal(Error{err.Error()})
		if err2 != nil {
			return err2
		}
		p.command.PrintErrln(string(result))
	}
	return err
}
