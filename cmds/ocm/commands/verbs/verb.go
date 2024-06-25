package verbs

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context, name string, short string) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: short,
	}, name)
	return cmd
}
