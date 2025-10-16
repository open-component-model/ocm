package components

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/add"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/check"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/download"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/get"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/hash"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/list"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/sign"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/transfer"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/verify"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Components

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on components",
	}, Names...)
	AddCommands(ctx, cmd)
	return cmd
}

func AddCommands(ctx clictx.Context, cmd *cobra.Command) {
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(list.NewCommand(ctx, list.Verb))
	cmd.AddCommand(hash.NewCommand(ctx, hash.Verb))
	cmd.AddCommand(sign.NewCommand(ctx, sign.Verb))
	cmd.AddCommand(transfer.NewCommand(ctx, transfer.Verb))
	cmd.AddCommand(verify.NewCommand(ctx, verify.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
	cmd.AddCommand(check.NewCommand(ctx, check.Verb))
}
