package components

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	ocmcomp "ocm.software/ocm/cmds/ocm/commands/ocmcmds/components"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	toicomp "ocm.software/ocm/cmds/ocm/commands/toicmds/package"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Components

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on components",
	}, Names...)
	ocmcomp.AddCommands(ctx, cmd)
	toicomp.AddCommands(ctx, cmd)
	return cmd
}
