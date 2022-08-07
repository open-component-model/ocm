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

package components

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/sign"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/verify"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
)

var Names = names.Components

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on components",
	}, Names...)
	AddCommands(ctx, cmd)
	return cmd
}

func AddCommands(ctx clictx.Context, cmd *cobra.Command) {
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(sign.NewCommand(ctx, sign.Verb))
	cmd.AddCommand(verify.NewCommand(ctx, verify.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
}
