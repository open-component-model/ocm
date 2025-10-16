package verified

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/verified/get"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Verified

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on verified component versions",
	}, Names...)
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	return cmd
}
