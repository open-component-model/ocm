package add

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	components "ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/add"
	references "ocm.software/ocm/cmds/ocm/commands/ocmcmds/references/add"
	resourceconfig "ocm.software/ocm/cmds/ocm/commands/ocmcmds/resourceconfig/add"
	resources "ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources/add"
	routingslips "ocm.software/ocm/cmds/ocm/commands/ocmcmds/routingslips/add"
	sourceconfig "ocm.software/ocm/cmds/ocm/commands/ocmcmds/sourceconfig/add"
	sources "ocm.software/ocm/cmds/ocm/commands/ocmcmds/sources/add"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Add elements to a component repository or component version",
	}, verbs.Add)
	cmd.AddCommand(resourceconfig.NewCommand(ctx))
	cmd.AddCommand(sourceconfig.NewCommand(ctx))

	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(routingslips.NewCommand(ctx))
	return cmd
}
