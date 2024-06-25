package verbs

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context, name string, short string) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: short,
	}, name)
	return cmd
}
