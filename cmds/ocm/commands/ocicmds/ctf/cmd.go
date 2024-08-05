package ctf

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/ctf/create"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.TransportArchive

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on OCI view of a Common Transport Archive",
	}, Names...)
	cmd.AddCommand(create.NewCommand(ctx, create.Verb))
	return cmd
}
