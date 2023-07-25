// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocmcmds

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/componentarchive"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/ctf"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/plugins"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/references"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/resourceconfig"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/resources"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/sourceconfig"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/sources"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/versions"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	topicocmaccessmethods "github.com/open-component-model/ocm/v2/cmds/ocm/topics/ocm/accessmethods"
	topicocmdownloaders "github.com/open-component-model/ocm/v2/cmds/ocm/topics/ocm/downloadhandlers"
	topicocmrefs "github.com/open-component-model/ocm/v2/cmds/ocm/topics/ocm/refs"
	topicocmuploaders "github.com/open-component-model/ocm/v2/cmds/ocm/topics/ocm/uploadhandlers"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
)

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Dedicated command flavors for the Open Component Model",
	}, "ocm")
	cmd.AddCommand(resourceconfig.NewCommand(ctx))
	cmd.AddCommand(sourceconfig.NewCommand(ctx))
	cmd.AddCommand(resources.NewCommand(ctx))
	cmd.AddCommand(sources.NewCommand(ctx))
	cmd.AddCommand(references.NewCommand(ctx))
	cmd.AddCommand(components.NewCommand(ctx))
	cmd.AddCommand(ctf.NewCommand(ctx))
	cmd.AddCommand(componentarchive.NewCommand(ctx))
	cmd.AddCommand(versions.NewCommand(ctx))
	cmd.AddCommand(plugins.NewCommand(ctx))

	cmd.AddCommand(topicocmrefs.New(ctx))
	cmd.AddCommand(topicocmaccessmethods.New(ctx))
	cmd.AddCommand(topicocmuploaders.New(ctx))
	cmd.AddCommand(topicocmdownloaders.New(ctx))

	return cmd
}
