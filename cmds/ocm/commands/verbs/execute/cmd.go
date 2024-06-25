package execute

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	action "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/action/execute"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Execute an element.",
	}, verbs.Execute)
	cmd.AddCommand(action.NewCommand(ctx))
	return cmd
}
