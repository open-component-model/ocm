package versions

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/versions/show"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Versions

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component version names",
	}, Names...)
	cmd.AddCommand(show.NewCommand(ctx, show.Verb))
	return cmd
}
