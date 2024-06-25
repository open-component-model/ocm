package show

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	tags "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/tags/show"
	versions "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/versions/show"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Show tags or versions",
	}, verbs.Show)
	cmd.AddCommand(versions.NewCommand(ctx))
	cmd.AddCommand(tags.NewCommand(ctx))

	return cmd
}
