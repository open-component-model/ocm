package config

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/config/bootstrap"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Configuration

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "TOI Commands acting on config",
	}, Names...)
	AddCommands(ctx, cmd)
	return cmd
}

func AddCommands(ctx clictx.Context, cmd *cobra.Command) {
	cmd.AddCommand(bootstrap.NewCommand(ctx, bootstrap.Verb))
}
