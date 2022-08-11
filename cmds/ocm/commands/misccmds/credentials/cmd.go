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

package credentials

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"

	credentials "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/credentials/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
)

var Names = names.Credentials

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on credentials",
	}, Names...)
	cmd.AddCommand(credentials.NewCommand(ctx, credentials.Verb))
	return cmd
}
