package resources

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources/add"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources/download"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources/get"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Resources

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component resources",
	}, Names...)
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
	return cmd
}
