package describe

import (
	"github.com/spf13/cobra"

	_package "github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/package/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "describe packages",
	}, verbs.Describe)
	cmd.AddCommand(_package.NewCommand(ctx))
	return cmd
}
