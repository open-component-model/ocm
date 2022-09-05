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

package ocicmds

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artefacts"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/ctf"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/tags"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	topicocirefs "github.com/open-component-model/ocm/cmds/ocm/topics/oci/refs"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the OCI layer",
	}, "oci")
	cmd.AddCommand(artefacts.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(tags.NewCommand(ctx))

	cmd.AddCommand(topicocirefs.New(ctx))
	return cmd
}
