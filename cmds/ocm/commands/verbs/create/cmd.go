package create

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	rsakeypair "ocm.software/ocm/cmds/ocm/commands/misccmds/rsakeypair"
	ctf "ocm.software/ocm/cmds/ocm/commands/ocicmds/ctf/create"
	comparch "ocm.software/ocm/cmds/ocm/commands/ocmcmds/componentarchive/create"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Create transport or component archive",
	}, verbs.Create)
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	cmd.AddCommand(comparch.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(rsakeypair.NewCommand(ctx))
	return cmd
}
