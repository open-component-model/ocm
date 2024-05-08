package controller

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/controllercmds/uninstall"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/controllercmds/install"
	"github.com/open-component-model/ocm/cmds/ocm/commands/controllercmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on the ocm-controller",
	}, names.Controller...)
	cmd.AddCommand(install.NewCommand(ctx, install.Verb))
	cmd.AddCommand(uninstall.NewCommand(ctx, uninstall.Verb))
	return cmd
}
