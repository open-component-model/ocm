// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/spf13/cobra"

	artifacts "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/transfer"
	comparch "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive/transfer"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/transfer"
	ctf "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/ctf/transfer"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Transfer artifacts or components",
	}, verbs.Transfer)
	cmd.AddCommand(comparch.NewCommand(ctx))
	cmd.AddCommand(artifacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))

	return cmd
}
