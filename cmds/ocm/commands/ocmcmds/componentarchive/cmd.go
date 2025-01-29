package componentarchive

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/componentarchive/create"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/componentarchive/transfer"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
//
//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var Names = names.ComponentArchive

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "(DEPRECATED) - Please use " + names.CommonTransportArchive[0] + " instead",
		// this removes the command from the help output - https://github.com/open-component-model/ocm/issues/1242#issuecomment-2609312927
		// Deprecated: "Deprecated - use " + ocm.CommonTransportFormat + " instead",
	}, Names...)
	cmd.AddCommand(transfer.NewCommand(ctx, transfer.Verb))
	cmd.AddCommand(create.NewCommand(ctx, create.Verb))
	return cmd
}
