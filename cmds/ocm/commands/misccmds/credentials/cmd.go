package credentials

import (
	"github.com/spf13/cobra"

	credentials "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/credentials/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
