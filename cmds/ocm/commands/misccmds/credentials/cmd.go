package credentials

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	credentials "ocm.software/ocm/cmds/ocm/commands/misccmds/credentials/get"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Credentials

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on credentials",
	}, Names...)
	cmd.AddCommand(credentials.NewCommand(ctx, credentials.Verb))
	return cmd
}
