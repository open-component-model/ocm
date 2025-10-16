package install

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	plugins "ocm.software/ocm/cmds/ocm/commands/ocmcmds/plugins/install"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Install new OCM CLI components ",
	}, verbs.Install)
	cmd.AddCommand(plugins.NewCommand(ctx))
	return cmd
}
