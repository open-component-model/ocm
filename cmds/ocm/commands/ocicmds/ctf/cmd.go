package ctf

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/ctf/create"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var Names = names.TransportArchive

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on OCI view of a Common Transport Archive",
	}, Names...)
	cmd.AddCommand(create.NewCommand(ctx, create.Verb))
	return cmd
}
