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

package sources

import (
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources/get"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/spf13/cobra"
)

var Names = names.Sources

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component sources",
	}, Names...)
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	return cmd
}
