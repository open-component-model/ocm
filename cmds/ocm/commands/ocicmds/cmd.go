package ocicmds

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/ctf"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/tags"
	"ocm.software/ocm/cmds/ocm/common/utils"
	topicocirefs "ocm.software/ocm/cmds/ocm/topics/oci/refs"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the OCI layer",
	}, "oci")
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(tags.NewCommand(ctx))

	cmd.AddCommand(utils.DocuCommandPath(topicocirefs.New(ctx), "ocm"))
	return cmd
}
