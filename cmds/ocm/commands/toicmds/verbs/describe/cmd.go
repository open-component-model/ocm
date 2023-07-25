// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"github.com/spf13/cobra"

	_package "github.com/open-component-model/ocm/v2/cmds/ocm/commands/toicmds/package/describe"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "describe packages",
	}, verbs.Describe)
	cmd.AddCommand(_package.NewCommand(ctx))
	return cmd
}
