package describe

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	_package "ocm.software/ocm/cmds/ocm/commands/toicmds/package/describe"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "describe packages",
	}, verbs.Describe)
	cmd.AddCommand(_package.NewCommand(ctx))
	return cmd
}
