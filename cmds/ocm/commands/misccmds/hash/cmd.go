package credentials

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/hash/sign"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

var Names = names.Hash

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on hashes",
	}, Names...)
	cmd.AddCommand(sign.NewCommand(ctx, sign.Verb))
	return cmd
}
