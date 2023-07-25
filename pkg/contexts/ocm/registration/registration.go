// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"github.com/open-component-model/ocm/v2/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/v2/pkg/contexts/datacontext/action/handlers"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	pluginaccess "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/accessmethods/plugin"
	pluginaction "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/actionhandler/plugin"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/plugincacheattr"
	pluginupload "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/blobhandler/handlers/generic/plugin"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/download"
	plugindownload "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/download/handlers/plugin"
	"github.com/open-component-model/ocm/v2/pkg/runtime"
)

// RegisterExtensions registers all the extension provided by the found plugin.
func RegisterExtensions(ctx ocm.Context) error {
	pi := plugincacheattr.Get(ctx)

	logger := Logger(ctx)
	for _, n := range pi.PluginNames() {
		p := pi.Get(n)
		if !p.IsValid() {
			continue
		}
		for _, a := range p.GetDescriptor().Actions {
			h, err := pluginaction.New(p, a.Name)
			if err != nil {
				logger.Error("cannot create action handler for plugin", "plugin", p.Name(), "handler", a.Name)
			} else {
				for _, s := range a.DefaultSelectors {
					err := ctx.AttributesContext().GetActions().Register(h, handlers.ForAction(a.Name), action.Selector(s))
					if err != nil {
						logger.LogError(err, "cannot register action handler for plugin", "plugin", p.Name(), "handler", a.Name, "selector", s)
					}
				}
			}
		}
		for _, m := range p.GetDescriptor().AccessMethods {
			name := m.Name
			if m.Version != "" {
				name = name + runtime.VersionSeparator + m.Version
			}
			logger.Info("registering access method",
				"plugin", p.Name(),
				"type", name)
			pi.GetContext().AccessMethods().Register(pluginaccess.NewType(name, p, &m))
		}

		if p.IsAutoConfigurationEnabled() {
			for _, u := range p.GetDescriptor().Uploaders {
				for _, c := range u.Constraints {
					if c.ContextType != "" && c.RepositoryType != "" && c.MediaType != "" {
						hdlr, err := pluginupload.New(p, u.Name, nil)
						if err != nil {
							logger.Error("cannot create blob handler for plugin", "plugin", p.Name(), "handler", u.Name)
						} else {
							logger.Info("registering repository blob handler",
								"context", c.ContextType+":"+c.RepositoryType,
								"plugin", p.Name(),
								"handler", u.Name)
							ctx.BlobHandlers().Register(hdlr, cpi.ForRepo(c.ContextType, c.RepositoryType), cpi.ForMimeType(c.MediaType))
						}
					}
				}
			}

			for _, u := range p.GetDescriptor().Downloaders {
				for _, c := range u.AutoRegistration {
					if c.ArtifactType != "" || c.MediaType != "" {
						hdlr, err := plugindownload.New(p, u.Name, nil)
						if err != nil {
							logger.Error("cannot create download handler for plugin", "plugin", p.Name(), "handler", u.Name)
						} else {
							logger.Info("registering download handler",
								"context", c.ArtifactType+":"+c.MediaType,
								"plugin", p.Name(),
								"handler", u.Name,
								"priority", c.Priority)
							opts := &download.HandlerOptions{
								HandlerKey: download.HandlerKey{
									ArtifactType: c.ArtifactType,
									MimeType:     c.MediaType,
								},
								Priority: c.Priority,
							}
							download.For(ctx).Register(hdlr, opts)
						}
					}
				}
			}
		}
	}
	return nil
}
