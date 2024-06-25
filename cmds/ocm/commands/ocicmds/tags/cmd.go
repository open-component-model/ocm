package tags

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/tags/show"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

var Names = names.Tags

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on OCI tag names",
	}, Names...)
	cmd.AddCommand(show.NewCommand(ctx, show.Verb))
	return cmd
}
