package versions

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/versions/show"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

var Names = names.Versions

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component version names",
	}, Names...)
	cmd.AddCommand(show.NewCommand(ctx, show.Verb))
	return cmd
}
