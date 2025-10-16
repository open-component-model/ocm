package toicmds

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/config"
	_package "ocm.software/ocm/cmds/ocm/commands/toicmds/package"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/verbs/bootstrap"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/verbs/describe"
	"ocm.software/ocm/cmds/ocm/common/utils"
	topicocmrefs "ocm.software/ocm/cmds/ocm/topics/ocm/refs"
	topicbootstrap "ocm.software/ocm/cmds/ocm/topics/toi/bootstrapping"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the TOI layer",
		Long: `
TOI is an abbreviation for Tiny OCM Installation. It is a simple
application framework on top of the Open Component Model, that can
be used to describe image based installation executors and installation
packages (see topic <CMD>ocm toi-bootstrapping</CMD> in form of resources
with a dedicated type. All involved resources are hereby taken from a component
version of the Open Component Model, which supports all the OCM features, like
transportation.

The framework consists of a generic bootstrap command
(<CMD>ocm bootstrap package</CMD>) and an arbitrary set of image
based executors, that are executed in containers and fed with the required
installation data by th generic command.
`,
	}, "toi")

	cmd.AddCommand(_package.NewCommand(ctx))
	cmd.AddCommand(config.NewCommand(ctx))

	cmd.AddCommand(bootstrap.NewCommand(ctx))
	cmd.AddCommand(describe.NewCommand(ctx))

	cmd.AddCommand(utils.DocuCommandPath(topicocmrefs.New(ctx), "ocm"))
	cmd.AddCommand(utils.DocuCommandPath(topicbootstrap.New(ctx, "bootstrapping"), "ocm", "toi-bootstrapping"))
	return cmd
}
