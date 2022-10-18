// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
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
	return utils.SetupCommand(
		&Command{
			common.ResourceAdderCommand{
				BaseCommand: utils.NewBaseCommand(ctx),
				Adder:       NewReferenceSpecificatonProvider(),
			},
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <target> {<referencefile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "add aggregation information to a component version",
		Long: `
Add aggregation information specified in a reference file to a component version.
So far only component archives are supported as target.

This command accepts reference specification files describing the references
to add to a component version. Elements must follow the reference meta data
description scheme of the component descriptor.
` + o.Adder.Description() + (&template.Options{}).Usage(),
		Example: `
Add a reference directly by options
<pre>
$ ocm add references path/to/ca --name myref --component github.com/my/component --version ${VERSION}
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
	return o.ProcessResourceDescriptions("references", ResourceSpecHandler{})
}
