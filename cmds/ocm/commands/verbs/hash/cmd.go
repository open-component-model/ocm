package hash

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/hash"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Hash and normalization operations",
	}, verbs.Hash)
	cmd.AddCommand(components.NewCommand(ctx))
	return cmd
}
