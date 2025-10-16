package sources

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/sources/add"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/sources/get"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Sources

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component sources",
	}, Names...)
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	return cmd
}
