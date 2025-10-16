package clean

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	cache "ocm.software/ocm/cmds/ocm/commands/cachecmds/clean"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Cleanup/re-organize elements",
	}, verbs.Clean)
	cmd.AddCommand(cache.NewCommand(ctx))
	return cmd
}
