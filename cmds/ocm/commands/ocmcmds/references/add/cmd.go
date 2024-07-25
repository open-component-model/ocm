package add

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/refs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/utils/template"
)

var (
	Names = names.References
	Verb  = verbs.Add
)

type Command struct {
	common.ResourceAdderCommand
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			common.NewResourceAdderCommand(ctx, refs.New().WithCLIOptions(&addhdlrs.Options{}), NewReferenceSpecificatonProvider()),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] [<target>] {<referencefile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(0),
		Short: "add aggregation information to a component version",
		Long: `
Add aggregation information specified in a reference file to a component version.
So far only component archives are supported as target.

This command accepts reference specification files describing the references
to add to a component version. Elements must follow the reference meta data
description scheme of the component descriptor.

The description file might contain:
- a single reference
- a list of references under the key <code>references</code>
- a list of yaml documents with a single reference or reference list

` + o.Adder.Description() + (&template.Options{}).Usage(),
		Example: `
Add a reference directly by options
<pre>
$ ocm add references --file path/to/ca --name myref --component github.com/my/component --version ${VERSION}
</pre>

Add a reference by a description file:

*references.yaml*:
<pre>
---
name: myref
component: github.com/my/component
version: ${VERSION]
</pre>
<pre>
$ ocm add references  path/to/ca  references.yaml VERSION=1.0.0
</pre>
`,
	}
}

func (o *Command) Run() error {
	return o.ProcessResourceDescriptions()
}
