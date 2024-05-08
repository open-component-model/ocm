package artifacts

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/transfer"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var Names = names.Artifacts

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on OCI artifacts",
	}, Names...)

	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(describe.NewCommand(ctx, describe.Verb))
	cmd.AddCommand(transfer.NewCommand(ctx, transfer.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
	return cmd
}
