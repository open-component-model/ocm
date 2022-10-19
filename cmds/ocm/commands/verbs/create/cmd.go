// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"github.com/spf13/cobra"

	rsakeypair "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/rsakeypair"
	ctf "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/ctf/create"
	comparch "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive/create"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Create transport or component archive",
	}, verbs.Create)
	cmd.AddCommand(comparch.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(rsakeypair.NewCommand(ctx))
	return cmd
}
