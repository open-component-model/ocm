package toicmds

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/config"
	_package "github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/package"
	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/verbs/bootstrap"
	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/verbs/describe"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	topicocmrefs "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/refs"
	topicbootstrap "github.com/open-component-model/ocm/cmds/ocm/topics/toi/bootstrapping"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the TOI layer",
		Long: `
TOI is an abbreviation for (T)iny (O)CM (I)nstallation. It is a simple
application framework on top of the Open Component Model, that can
be used to describe image based installation executors and installation
packages (see topic <CMD>ocm toi bootstrapping</CMD> in form of resources
with a dedicated type. All involved resources are hereby taken from a component
version of the Open Component Model, which supports all the OCM features, like
transportation.

The framework consists of a generic bootstrap command
(<CMD>ocm toi bootstrap componentversions</CMD>) and an arbitrary set of image
based executors, that are executed in containers and fed with the required
installation data by th generic command.
`,
	}, "toi")

	cmd.AddCommand(_package.NewCommand(ctx))
	cmd.AddCommand(config.NewCommand(ctx))

	cmd.AddCommand(bootstrap.NewCommand(ctx))
	cmd.AddCommand(describe.NewCommand(ctx))

	cmd.AddCommand(topicocmrefs.New(ctx))
	cmd.AddCommand(topicbootstrap.New(ctx, "bootstrapping"))
	return cmd
}
