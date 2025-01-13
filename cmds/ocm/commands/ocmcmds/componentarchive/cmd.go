package componentarchive

//nolint:staticcheck

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/componentarchive/create"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/componentarchive/transfer"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var Names = names.ComponentArchive

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component archives",
	}, Names...)
	cmd.AddCommand(transfer.NewCommand(ctx, transfer.Verb))
	cmd.AddCommand(create.NewCommand(ctx, create.Verb))
	return cmd
}
