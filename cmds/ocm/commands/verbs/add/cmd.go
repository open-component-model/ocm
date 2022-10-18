// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"github.com/spf13/cobra"

	references "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references/add"
	resourceconfig "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resourceconfig/add"
	resources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/add"
	sourceconfig "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sourceconfig/add"
	sources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Add resources or sources to a component archive",
	}, verbs.Add)
	cmd.AddCommand(resourceconfig.NewCommand(ctx))
	cmd.AddCommand(sourceconfig.NewCommand(ctx))

	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	return cmd
}
