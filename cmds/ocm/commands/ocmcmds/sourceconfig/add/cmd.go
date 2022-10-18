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
	rscadd "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/template"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

var (
	Names = names.SourceConfig
	Verb  = verbs.Add
)

type Command struct {
	common.ResourceConfigAdderCommand
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			common.ResourceConfigAdderCommand{
				BaseCommand: utils.NewBaseCommand(ctx),
				Adder:       common.NewContentResourceSpecificationProvider("source", resourcetypes.FILESYSTEM),
				Templating: template.Options{
					Default: "none",
				},
			},
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <target> {<configfile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "add a source specification to a source config file",
		Long: `
Add a source specification to a source config file used by <CMD>ocm add sources</CMD>.
` + o.Adder.Description() + ` Elements must follow the resource meta data
description scheme of the component descriptor.

If not specified anywhere the artifact type will be defaulted to <code>` + resourcetypes.FILESYSTEM + `</code>.

If expressions/templates are used in the specification file an appropriate
templater and the required settings might be required to provide
a correct input validation.

This command accepts additional source specification files describing the sources
to add to a component version.

` + (&template.Options{}).Usage() + inputs.Usage(inputs.DefaultInputTypeScheme),
		Example: `
$ ocm add source-config sources.yaml --name sources --type filesystem --access '{ "type": "gitHub", "repoUrl": "github.com/open-component-model/ocm", "commit": "xyz" }'
`,
	}
}

func (o *Command) Run() error {
	return o.ProcessResourceDescriptions("sources", rscadd.ResourceSpecHandler{})
}
