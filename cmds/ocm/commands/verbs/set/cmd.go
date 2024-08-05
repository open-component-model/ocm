package set

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	pubsub "ocm.software/ocm/cmds/ocm/commands/ocmcmds/pubsub/set"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new set command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Set information about OCM repositories",
	}, verbs.Set)
	cmd.AddCommand(pubsub.NewCommand(ctx))
	return cmd
}
