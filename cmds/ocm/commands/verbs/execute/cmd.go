package execute

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	action "ocm.software/ocm/cmds/ocm/commands/misccmds/action/execute"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Execute an element.",
	}, verbs.Execute)
	cmd.AddCommand(action.NewCommand(ctx))
	return cmd
}
