package transfer

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	artifacts "ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/transfer"
	comparch "ocm.software/ocm/cmds/ocm/commands/ocmcmds/componentarchive/transfer"
	components "ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/transfer"
	ctf "ocm.software/ocm/cmds/ocm/commands/ocmcmds/ctf/transfer"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Transfer artifacts or components",
	}, verbs.Transfer)
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	cmd.AddCommand(comparch.NewCommand(ctx))
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))

	return cmd
}
