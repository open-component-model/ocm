package pubsub

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/pubsub/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/pubsub/set"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
