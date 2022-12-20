// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/hash"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/sign"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/verify"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(hash.NewCommand(ctx, hash.Verb))
	cmd.AddCommand(sign.NewCommand(ctx, sign.Verb))
	cmd.AddCommand(verify.NewCommand(ctx, verify.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
}
