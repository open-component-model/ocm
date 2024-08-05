package hash

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	components "ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/hash"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Hash and normalization operations",
	}, verbs.Hash)
	cmd.AddCommand(components.NewCommand(ctx))
	return cmd
}
