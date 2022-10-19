// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
	return utils.SetupCommand(&Command{common.ResourceAdderCommand{BaseCommand: utils.NewBaseCommand(ctx)}}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <target> {<resourcefile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(2),
		Short: "add aggregation information to a component version",
		Long: `
Add aggregation information specified in a resource file to a component version.
So far only component archives are supported as target.
` + (&template.Options{}).Usage() + `
This command accepts reference specification files describing the references
to add to a component version.
` + inputs.Usage(inputs.DefaultInputTypeScheme),
	}
}

func (o *Command) Run() error {
	return o.ProcessResourceDescriptions("references", ResourceSpecHandler{})
}
