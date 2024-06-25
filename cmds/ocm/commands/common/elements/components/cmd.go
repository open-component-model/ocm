package components

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	ocmcomp "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	toicomp "github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/package"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

var Names = names.Components

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on components",
	}, Names...)
	ocmcomp.AddCommands(ctx, cmd)
	toicomp.AddCommands(ctx, cmd)
	return cmd
}
