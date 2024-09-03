package get

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	config "ocm.software/ocm/cmds/ocm/commands/misccmds/config/get"
	credentials "ocm.software/ocm/cmds/ocm/commands/misccmds/credentials/get"
	artifacts "ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/get"
	components "ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/get"
	plugins "ocm.software/ocm/cmds/ocm/commands/ocmcmds/plugins/get"
	pubsub "ocm.software/ocm/cmds/ocm/commands/ocmcmds/pubsub/get"
	references "ocm.software/ocm/cmds/ocm/commands/ocmcmds/references/get"
	resources "ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources/get"
	routingslips "ocm.software/ocm/cmds/ocm/commands/ocmcmds/routingslips/get"
	sources "ocm.software/ocm/cmds/ocm/commands/ocmcmds/sources/get"
	verified "ocm.software/ocm/cmds/ocm/commands/ocmcmds/verified/get"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Get information about artifacts and components",
	}, verbs.Get)
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(credentials.NewCommand(ctx))
	cmd.AddCommand(plugins.NewCommand(ctx))
	cmd.AddCommand(routingslips.NewCommand(ctx))
	cmd.AddCommand(config.NewCommand(ctx))
	cmd.AddCommand(pubsub.NewCommand(ctx))
	cmd.AddCommand(verified.NewCommand(ctx))
	return cmd
}
