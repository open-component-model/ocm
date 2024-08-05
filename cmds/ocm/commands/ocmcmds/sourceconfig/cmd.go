package sourceconfig

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/sourceconfig/add"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.SourceConfig

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component source specifications",
	}, Names...)
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	return cmd
}
