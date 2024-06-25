package clean

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	cache "github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/clean"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Cleanup/re-organize elements",
	}, verbs.Clean)
	cmd.AddCommand(cache.NewCommand(ctx))
	return cmd
}
