package verbs

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context, name string, short string) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: short,
	}, name)
	return cmd
}
