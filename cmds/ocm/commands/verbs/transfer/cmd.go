package transfer

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	artifacts "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/transfer"
	comparch "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive/transfer"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/transfer"
	ctf "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/ctf/transfer"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Transfer artifacts or components",
	}, verbs.Transfer)
	cmd.AddCommand(comparch.NewCommand(ctx))
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))

	return cmd
}
