// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package downloaderoption

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/optutils"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers/dirtree"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Registration = optutils.Registration

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New() *Option {
	return &Option{optutils.NewRegistrationOption("downloader", "", "artifact downloader", usage)}
}

type Option struct {
	optutils.RegistrationOption
}

func (o *Option) Register(ctx ocm.ContextProvider) error {
	for _, s := range o.Registrations {
		err := download.RegisterHandlerByName(ctx.OCMContext(), s.Name, s.Config,
			download.ForArtifactType(s.ArtifactType), download.ForMimeType(s.MediaType))
		if err != nil {
			return err
		}
	}
	return nil
}

var usage = `
- <code>ocm/dirtree</code>: downloading directory tree like resources.
  The following artifact media types are supported:
` + utils.IndentLines(listformat.FormatList("", dirtree.SupportedMimeTypes()...), "  ") + `
- <code>plugin/<plugin name>[/<downloader name]</code>: downloader provided by plugin.
`
