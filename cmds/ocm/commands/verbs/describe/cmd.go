// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"github.com/spf13/cobra"

	resources "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artifacts/describe"
	plugins "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Describe various elements by using appropriate sub commands.",
	}, verbs.Describe)
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(plugins.NewCommand(ctx))
	return cmd
}
