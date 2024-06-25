package create

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	rsakeypair "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/rsakeypair"
	ctf "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/ctf/create"
	comparch "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive/create"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Create transport or component archive",
	}, verbs.Create)
	cmd.AddCommand(comparch.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(rsakeypair.NewCommand(ctx))
	return cmd
}
