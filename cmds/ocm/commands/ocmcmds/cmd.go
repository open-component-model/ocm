// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocmcmds

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/ctf"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/versions"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	topicocmrefs "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/refs"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the Open Component Model",
	}, "ocm")
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(componentarchive.NewCommand(ctx))
	cmd.AddCommand(versions.NewCommand(ctx))

	cmd.AddCommand(topicocmrefs.New(ctx))
	return cmd
}
