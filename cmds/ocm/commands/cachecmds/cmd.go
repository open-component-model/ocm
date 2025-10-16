package cachecmds

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/cachecmds/clean"
	"ocm.software/ocm/cmds/ocm/commands/cachecmds/describe"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new cache command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Cache related commands",
	}, "cache")
	cmd.AddCommand(clean.NewCommand(ctx, clean.Verb))
	cmd.AddCommand(describe.NewCommand(ctx, describe.Verb))
	return cmd
}
