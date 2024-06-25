package install

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	plugins "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/install"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Install elements.",
	}, verbs.Install)
	cmd.AddCommand(plugins.NewCommand(ctx))
	return cmd
}
