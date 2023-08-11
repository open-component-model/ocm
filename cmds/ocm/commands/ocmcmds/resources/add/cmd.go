// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs/rscs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

var (
	Names = names.Resources
	Verb  = verbs.Add
)

type Command struct {
	common.ResourceAdderCommand
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			common.NewResourceAdderCommand(ctx, rscs.New(), NewResourceSpecificationsProvider(ctx, "")),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] [<target>] {<resourcefile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(0),
		Short: "add resources to a component version",
		Example: `
Add a resource directly by options
<pre>
$ ocm add resources --file path/to/ca --name myresource --type PlainText --input '{ "type": "file", "path": "testdata/testcontent", "mediaType": "text/plain" }'
</pre>

Add a resource by a description file:

*resources.yaml*:
<pre>
---
name: myrresource
type: PlainText
version: ${version]
input:
  type: file
  path: testdata/testcontent
  mediaType: text/plain
</pre>
<pre>
$ ocm add resources --file path/to/ca  resources.yaml VERSION=1.0.0
</pre>
`,
	}
}

func (o *Command) Long() string {
	return `
Add resources specified in a resource file to a component version.
So far only component archives are supported as target.

This command accepts resource specification files describing the resources
to add to a component version. Elements must follow the resource meta data
description scheme of the component descriptor. Besides referential resources
using the <code>access</code> attribute to describe the access method, it
is possible to describe local resources fed by local data using the <code>input</code>
field (see below).

The description file might contain:
- a single resource
- a list of resources under the key <code>resources</code>
- a list of yaml documents with a single resource or resource list

` + o.Adder.Description() + (&template.Options{}).Usage() +
		inputs.Usage(inputs.DefaultInputTypeScheme) +
		ocm.AccessUsage(o.OCMContext().AccessMethods(), true)
}

func (o *Command) Run() error {
	return o.ProcessResourceDescriptions()
}
