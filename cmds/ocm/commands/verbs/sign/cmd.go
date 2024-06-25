package sign

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/hash/sign"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/sign"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
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
