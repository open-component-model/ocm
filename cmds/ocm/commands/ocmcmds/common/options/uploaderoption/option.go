// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package uploaderoption

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/optutils"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler"
)

type Registration = optutils.Registration

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New() *Option {
	return &Option{optutils.NewRegistrationOption("uploader", "", "repository uploader", usage)}
}

type Option struct {
	optutils.RegistrationOption
}

func (o *Option) Register(ctx ocm.ContextProvider) error {
	for _, s := range o.Registrations {
		err := blobhandler.RegisterHandlerByName(ctx.OCMContext(), s.Name, s.Config,
			blobhandler.ForArtifactType(s.ArtifactType), blobhandler.ForMimeType(s.MediaType))
		if err != nil {
			return err
		}
	}
	return nil
}

const usage = `
- <code>ocm/ociRegistry</code>: oci Registry upload for local OCI artifact blobs.
  The media type is optional. If given ist must be an OCI artifact media type.
- <code>plugin/<plugin name>[/<uploader name]</code>: uploader provided by plugin.
`
