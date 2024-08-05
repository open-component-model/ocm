package ocmcmds

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/componentarchive"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/components"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/ctf"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/plugins"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/pubsub"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/references"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/resourceconfig"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/resources"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/routingslips"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/sourceconfig"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/sources"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/versions"
	"ocm.software/ocm/cmds/ocm/common/utils"
	topicocmaccessmethods "ocm.software/ocm/cmds/ocm/topics/ocm/accessmethods"
	topicocmdownloaders "ocm.software/ocm/cmds/ocm/topics/ocm/downloadhandlers"
	topicocmrefs "ocm.software/ocm/cmds/ocm/topics/ocm/refs"
	topicocmuploaders "ocm.software/ocm/cmds/ocm/topics/ocm/uploadhandlers"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the Open Component Model",
	}, "ocm")
	cmd.AddCommand(resourceconfig.NewCommand(ctx))
	cmd.AddCommand(sourceconfig.NewCommand(ctx))
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(componentarchive.NewCommand(ctx))
	cmd.AddCommand(versions.NewCommand(ctx))
	cmd.AddCommand(plugins.NewCommand(ctx))
	cmd.AddCommand(routingslips.NewCommand(ctx))
	cmd.AddCommand(pubsub.NewCommand(ctx))

	cmd.AddCommand(topicocmrefs.New(ctx))
	cmd.AddCommand(topicocmaccessmethods.New(ctx))
	cmd.AddCommand(topicocmuploaders.New(ctx))
	cmd.AddCommand(topicocmdownloaders.New(ctx))

	return cmd
}
