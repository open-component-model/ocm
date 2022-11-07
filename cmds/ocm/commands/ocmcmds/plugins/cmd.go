// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/install"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var Names = names.Plugins

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands related to OCM plugins",
	}, Names...)
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(install.NewCommand(ctx, install.Verb))
	cmd.AddCommand(describe.NewCommand(ctx, describe.Verb))
	return cmd
}
