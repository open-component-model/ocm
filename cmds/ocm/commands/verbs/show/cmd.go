package show

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	tags "ocm.software/ocm/cmds/ocm/commands/ocicmds/tags/show"
	versions "ocm.software/ocm/cmds/ocm/commands/ocmcmds/versions/show"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Show tags or versions",
	}, verbs.Show)
	cmd.AddCommand(versions.NewCommand(ctx))
	cmd.AddCommand(tags.NewCommand(ctx))

	return cmd
}
