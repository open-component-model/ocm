package ocicmds

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/ctf"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/tags"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	topicocirefs "github.com/open-component-model/ocm/cmds/ocm/topics/oci/refs"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the OCI layer",
	}, "oci")
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(tags.NewCommand(ctx))

	cmd.AddCommand(topicocirefs.New(ctx))
	return cmd
}
