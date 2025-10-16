package describe

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	cache "ocm.software/ocm/cmds/ocm/commands/cachecmds/describe"
	resources "ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/describe"
	plugins "ocm.software/ocm/cmds/ocm/commands/ocmcmds/plugins/describe"
	_package "ocm.software/ocm/cmds/ocm/commands/toicmds/package/describe"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Describe various elements by using appropriate sub commands.",
	}, verbs.Describe)
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(plugins.NewCommand(ctx))
	cmd.AddCommand(cache.NewCommand(ctx))
	cmd.AddCommand(_package.NewCommand(ctx))
	return cmd
}
