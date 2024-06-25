package credentials

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	config "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/config/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

var Names = names.Config

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on CLI config",
	}, Names...)
	cmd.AddCommand(config.NewCommand(ctx, config.Verb))
	return cmd
}
