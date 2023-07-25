// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cachecmds

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/cachecmds/clean"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/cachecmds/info"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

// NewCommand creates a new cache command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Cache related commands",
	}, "cache")
	cmd.AddCommand(clean.NewCommand(ctx, clean.Verb))
	cmd.AddCommand(info.NewCommand(ctx, info.Verb))
	return cmd
}
