package cli

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/cli/download"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Components

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on the command line client (ocm)",
	}, Names...)
	AddCommands(ctx, cmd)
	return cmd
}

func AddCommands(ctx clictx.Context, cmd *cobra.Command) {
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
}
