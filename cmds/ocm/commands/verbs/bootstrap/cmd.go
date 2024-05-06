package bootstrap

import (
	"github.com/spf13/cobra"

	config "github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/config/bootstrap"
	_package "github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/package/bootstrap"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "bootstrap components",
	}, verbs.Bootstrap)
	cmd.AddCommand(_package.NewCommand(ctx))
	cmd.AddCommand(config.NewCommand(ctx))
	return cmd
}
