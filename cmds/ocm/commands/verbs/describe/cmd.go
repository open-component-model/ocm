package describe

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	cache "github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/describe"
	resources "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/describe"
	plugins "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/describe"
	_package "github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/package/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
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
