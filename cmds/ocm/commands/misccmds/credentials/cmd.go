package credentials

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	credentials "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/credentials/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

var Names = names.Credentials

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on credentials",
	}, Names...)
	cmd.AddCommand(credentials.NewCommand(ctx, credentials.Verb))
	return cmd
}
