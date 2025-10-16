package pubsub

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/pubsub/get"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/pubsub/set"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.PubSub

// NewCommand creates a new pubsub command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on sub/sub specifications",
	}, Names...)
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(set.NewCommand(ctx, set.Verb))
	return cmd
}
