package sign

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/hash/sign"
	components "ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/sign"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Sign components or hashes",
	}, verbs.Sign)
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(sign.NewCommand(ctx))
	return cmd
}
