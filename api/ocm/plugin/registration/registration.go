package registration

import (
	"slices"

	"ocm.software/ocm/api/config/plugin"
	"ocm.software/ocm/api/datacontext/action"
	"ocm.software/ocm/api/datacontext/action/handlers"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	pluginaccess "ocm.software/ocm/api/ocm/extensions/accessmethods/plugin"
	pluginaction "ocm.software/ocm/api/ocm/extensions/actionhandler/plugin"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	pluginupload "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/generic/plugin"
	"ocm.software/ocm/api/ocm/extensions/download"
	plugindownload "ocm.software/ocm/api/ocm/extensions/download/handlers/plugin"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/spi"
	pluginroutingslip "ocm.software/ocm/api/ocm/extensions/labels/routingslip/types/plugin"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/valuemergehandler"
	pluginmerge "ocm.software/ocm/api/ocm/valuemergehandler/handlers/plugin"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
	"ocm.software/ocm/api/utils/runtime"
)

// RegisterExtensions registers all the extension provided by the found plugin.
func RegisterExtensions(ctxp ocm.ContextProvider) error {
	ctx := ctxp.OCMContext()
	pi := plugincacheattr.Get(ctx)

	logger := Logger(ctx)
	vmreg := valuemergehandler.For(ctx)
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

		for _, a := range p.GetDescriptor().ValueMergeHandlers {
			h, err := pluginmerge.New(p, a.Name)
			if err != nil {
				logger.Error("cannot create value merge handler for plugin", "plugin", p.Name(), "handler", a.Name)
			} else {
				vmreg.RegisterHandler(h)
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

		for _, m := range p.GetDescriptor().ValueSets {
			if !slices.Contains(m.Purposes, descriptor.PURPOSE_ROUTINGSLIP) {
				continue
			}
			name := m.Name
			if m.Version != "" {
				name = name + runtime.VersionSeparator + m.Version
			}
			logger.Info("registering routing slip entry type",
				"plugin", p.Name(),
				"type", name)
			spi.For(pi.GetContext()).Register(pluginroutingslip.NewType(name, p, &m))
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

		for _, s := range p.GetDescriptor().LabelMergeSpecifications {
			h := vmreg.GetHandler(s.GetAlgorithm())
			if h == nil {
				logger.Error("cannot assign label merge spec for plugin", "label", s.GetName(), "algorithm", s.GetAlgorithm(), "plugin", p.Name())
			} else {
				vmreg.AssignHandler(hpi.LabelHint(s.Name, s.Version), &s.MergeAlgorithmSpecification)
			}
		}

		registry := ctx.ConfigContext().ConfigTypes()
		for _, s := range p.GetDescriptor().ConfigTypes {
			name := s.Name
			if s.Version != "" {
				name += runtime.VersionSeparator + s.Version
			}
			if registry.GetType(name) != nil {
				logger.Error("config type {{type}} already registered", "type", name)
			}
			t := plugin.New(name, s.Description)
			registry.Register(t)
		}
	}
	return nil
}
