// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package clean

import (
	"github.com/spf13/cobra"

	cache "github.com/open-component-model/ocm/v2/cmds/ocm/commands/cachecmds/clean"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Cleanup/re-organize elements",
	}, verbs.Clean)
	cmd.AddCommand(cache.NewCommand(ctx))
	return cmd
}
