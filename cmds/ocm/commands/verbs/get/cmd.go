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

package get

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/spf13/cobra"

	credentials "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/credentials/get"
	artefacts "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artefacts/get"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/get"
	references "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references/get"
	resources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/get"
	sources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources/get"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Get information about artefacts and components",
	}, verbs.Get)
	cmd.AddCommand(artefacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(credentials.NewCommand(ctx))
	return cmd
}
