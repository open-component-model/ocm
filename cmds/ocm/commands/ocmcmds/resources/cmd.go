// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/resources/add"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/resources/download"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/resources/get"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

var Names = names.Resources

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component resources",
	}, Names...)
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
	return cmd
}
