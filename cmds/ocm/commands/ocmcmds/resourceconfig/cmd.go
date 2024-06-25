package resourceconfig

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resourceconfig/add"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

var Names = names.ResourceConfig

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component resource specifications",
	}, Names...)
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	return cmd
}
