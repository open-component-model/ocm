// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package artefacts

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artefacts/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artefacts/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artefacts/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artefacts/transfer"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var Names = names.Artefacts

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on OCI artefacts",
	}, Names...)

	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(describe.NewCommand(ctx, describe.Verb))
	cmd.AddCommand(transfer.NewCommand(ctx, transfer.Verb))
	cmd.AddCommand(download.NewCommand(ctx, download.Verb))
	return cmd
}
