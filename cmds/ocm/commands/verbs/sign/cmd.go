// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package sign

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/hash/sign"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/sign"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Sign components or hashes",
	}, verbs.Sign)
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(sign.NewCommand(ctx))
	return cmd
}
