package hash

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/hash/sign"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/names"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.Hash

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on hashes",
	}, Names...)
	cmd.AddCommand(sign.NewCommand(ctx, sign.Verb))
	return cmd
}
