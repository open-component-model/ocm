package plugins

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/plugins/describe"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/plugins/get"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/plugins/install"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Plugins

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands related to OCM plugins",
	}, Names...)
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(install.NewCommand(ctx, install.Verb))
	cmd.AddCommand(describe.NewCommand(ctx, describe.Verb))
	return cmd
}
