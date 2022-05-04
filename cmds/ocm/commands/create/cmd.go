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

package create

import (
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"

	ctf "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/ctf/create"
	comparch "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive/create"
	"github.com/spf13/cobra"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Create transport or component archive",
	}, commands.Create)
	cmd.AddCommand(comparch.NewCommand(ctx, comparch.Names...))
	cmd.AddCommand(ctf.NewCommand(ctx, ctf.Names...))
	return cmd
}
