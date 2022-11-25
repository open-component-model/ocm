// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"github.com/spf13/cobra"

	credentials "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/credentials/get"
	artifacts "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/get"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/get"
	plugins "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/get"
	references "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references/get"
	resources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/get"
	sources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Get information about artifacts and components",
	}, verbs.Get)
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(credentials.NewCommand(ctx))
	cmd.AddCommand(plugins.NewCommand(ctx))
	return cmd
}
