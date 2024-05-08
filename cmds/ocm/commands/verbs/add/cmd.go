package add

import (
	"github.com/spf13/cobra"

	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/add"
	references "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references/add"
	resourceconfig "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resourceconfig/add"
	resources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/add"
	routingslips "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/routingslips/add"
	sourceconfig "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sourceconfig/add"
	sources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
