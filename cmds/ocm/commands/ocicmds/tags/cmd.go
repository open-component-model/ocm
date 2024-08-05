package tags

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/tags/show"
	"ocm.software/ocm/cmds/ocm/common/utils"
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
