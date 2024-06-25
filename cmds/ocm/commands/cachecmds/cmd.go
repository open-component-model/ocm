package cachecmds

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/clean"
	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds/describe"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new cache command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Cache related commands",
	}, "cache")
	cmd.AddCommand(clean.NewCommand(ctx, clean.Verb))
	cmd.AddCommand(describe.NewCommand(ctx, describe.Verb))
	return cmd
}
