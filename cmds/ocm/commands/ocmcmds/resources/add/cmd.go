// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
			common.ResourceAdderCommand{
				BaseCommand: utils.NewBaseCommand(ctx),
				Adder:       NewResourceSpecificationsProvider(ctx, ""),
			},
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <target> {<resourcefile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "add resources to a component version",
		Long: `
Add resources specified in a resource file to a component version.
So far only component archives are supported as target.

This command accepts  resource specification files describing the resources
to add to a component version. Elements must follow the resource meta data
description scheme of the component descriptor.
` + o.Adder.Description() + (&template.Options{}).Usage() + inputs.Usage(inputs.DefaultInputTypeScheme),
		Example: `
Add a resource directly by options
<pre>
$ ocm add resources path/to/ca --name myresource --type PlainText --input '{ "type": "file", "path": "testdata/testcontent", "mediaType": "text/plain" }'
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
$ ocm add resources  path/to/ca  resources.yaml VERSION=1.0.0
</pre>
`,
	}
}

func (o *Command) Run() error {
	return o.ProcessResourceDescriptions("resources", ResourceSpecHandler{})
}
