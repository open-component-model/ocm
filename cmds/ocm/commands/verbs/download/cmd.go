// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"github.com/spf13/cobra"

	artefacts "github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/artefacts/download"
	components "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components/download"
	resources "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Download oci artefacts, resources or complete components",
	}, verbs.Download)
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(artefacts.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	return cmd
}
