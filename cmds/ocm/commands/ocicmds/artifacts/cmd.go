package artifacts

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/describe"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/download"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/get"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/transfer"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Artifacts

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on OCI artifacts",
	}, Names...)

	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(describe.NewCommand(ctx, describe.Verb))
	cmd.AddCommand(transfer.NewCommand(ctx, transfer.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
	return cmd
}
