// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"github.com/spf13/cobra"

	ocmcomp "github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/names"
	toicomp "github.com/open-component-model/ocm/v2/cmds/ocm/commands/toicmds/package"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

var Names = names.Components

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on components",
	}, Names...)
	ocmcomp.AddCommands(ctx, cmd)
	toicomp.AddCommands(ctx, cmd)
	return cmd
}
