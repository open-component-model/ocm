package add

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils/template"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/srcs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Sources
	Verb  = verbs.Add
)

type Command struct {
	common.ResourceAdderCommand
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			common.NewResourceAdderCommand(ctx, srcs.New().WithCLIOptions(&addhdlrs.Options{}), common.NewContentResourceSpecificationProvider(ctx, "source", nil, "")),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] [<target>] {<resourcefile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(0),
		Short: "add source information to a component version",
		Example: `
$ ocm add sources --file path/to/cafile sources.yaml
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) Long() string {
	return `
Add information about the sources, e.g. commits in a Github repository,
that have been used to create the resources specified in a resource file to a component version.
So far only component archives are supported as target.

This command accepts source specification files describing the sources
to add to a component version. Elements must follow the source meta data
description scheme of the component descriptor. Besides referential sources
using the <code>access</code> attribute to describe the access method, it
is possible to describe local sources fed by local data using the <code>input</code>
field (see below).

The description file might contain:
- a single source
- a list of sources under the key <code>sources</code>
- a list of yaml documents with a single source or source list

` + o.Adder.Description() + (&template.Options{}).Usage() +
		inputs.Usage(inputs.DefaultInputTypeScheme) +
		ocm.AccessUsage(o.OCMContext().AccessMethods(), true) + `

` + (&addhdlrs.Options{}).Description()
}

func (o *Command) Run() error {
	return o.ProcessResourceDescriptions()
}
