package install

import (
	"github.com/spf13/cobra"

	plugins "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/install"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Install elements.",
	}, verbs.Install)
	cmd.AddCommand(plugins.NewCommand(ctx))
	return cmd
}
