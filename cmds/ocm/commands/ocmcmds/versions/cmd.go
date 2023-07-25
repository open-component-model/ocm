// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package versions

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/versions/show"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

var Names = names.Versions

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands acting on component version names",
	}, Names...)
	cmd.AddCommand(show.NewCommand(ctx, show.Verb))
	return cmd
}
