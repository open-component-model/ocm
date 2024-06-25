package bootstrap

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/package/bootstrap"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "bootstrap components",
	}, verbs.Bootstrap)
	cmd.AddCommand(components.NewCommand(ctx))
	return cmd
}
