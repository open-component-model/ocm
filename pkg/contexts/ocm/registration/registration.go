// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/handlers"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/plugin"
	plugin2 "github.com/open-component-model/ocm/pkg/contexts/ocm/actionhandler/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/generic/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
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
			h, err := plugin2.New(p, a.Name)
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
			pi.GetContext().AccessMethods().Register(access.NewType(name, p, &m))
		}

		for _, u := range p.GetDescriptor().Uploaders {
			for _, c := range u.Constraints {
				if c.ContextType != "" && c.RepositoryType != "" && c.MediaType != "" {
					hdlr, err := plugin.New(p, u.Name, nil)
					if err != nil {
						logger.Error("cannot create blob handler fpr plugin", "plugin", p.Name(), "handler", u.Name)
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
	}
	return nil
}
