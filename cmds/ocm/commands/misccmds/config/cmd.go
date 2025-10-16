package config

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	config "ocm.software/ocm/cmds/ocm/commands/misccmds/config/get"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
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
