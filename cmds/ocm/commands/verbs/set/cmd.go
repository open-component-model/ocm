package set

import (
	"github.com/spf13/cobra"

	pubsub "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/pubsub/set"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new set command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Set information about OCM repositories",
	}, verbs.Set)
	cmd.AddCommand(pubsub.NewCommand(ctx))
	return cmd
}
