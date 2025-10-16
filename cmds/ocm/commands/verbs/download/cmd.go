package download

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	artifacts "ocm.software/ocm/cmds/ocm/commands/ocicmds/artifacts/download"
	cli "ocm.software/ocm/cmds/ocm/commands/ocmcmds/cli/download"
	components "ocm.software/ocm/cmds/ocm/commands/ocmcmds/components/download"
	resources "ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources/download"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Download oci artifacts, resources or complete components",
	}, verbs.Download)
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(cli.NewCommand(ctx))
	return cmd
}
