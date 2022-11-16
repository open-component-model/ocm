// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"encoding/json"
	"sync"

	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/plugin"
	blob "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/generic/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Set = *pluginsImpl

type pluginsImpl struct {
	lock sync.RWMutex

	updater cfgcpi.Updater
	ctx     ocm.Context
	base    cache.PluginDir
	configs map[string]json.RawMessage
	plugins map[string]plugin.Plugin
}

var _ config.Target = (*pluginsImpl)(nil)

func New(ctx ocm.Context, path string) Set {
	pi := &pluginsImpl{
		ctx:     ctx,
		configs: map[string]json.RawMessage{},
		plugins: map[string]plugin.Plugin{},
	}
	pi.updater = cfgcpi.NewUpdater(ctx.ConfigContext(), pi)
	pi.Update()
	pi.base = cache.Get(path)
	for _, n := range pi.base.PluginNames() {
		pi.plugins[n] = plugin.NewPlugin(ctx, pi.base.Get(n), pi.configs[n])
	}
	return pi
}

func (pi *pluginsImpl) Update() {
	err := pi.updater.Update()
	if err != nil {
		pi.ctx.Logger(internal.TAG).Error("config update failed", "error", err.Error())
	}
}

func (pi *pluginsImpl) ConfigurePlugin(name string, config json.RawMessage) {
	pi.lock.Lock()
	defer pi.lock.Unlock()

	pi.configs[name] = config
	if pi.plugins[name] != nil {
		pi.plugins[name].SetConfig(config)
	}
}

func (pi *pluginsImpl) PluginNames() []string {
	pi.lock.RLock()
	defer pi.lock.RUnlock()

	return utils.StringMapKeys(pi.plugins)
}

func (pi *pluginsImpl) Get(name string) plugin.Plugin {
	pi.Update()

	pi.lock.RLock()
	defer pi.lock.RUnlock()

	p, ok := pi.plugins[name]
	if ok {
		return p
	}
	return nil
}

// RegisterExtensions registers all the extension provided the found plugin
// at the given context. If no context is given, the cache context is used.
func (pi *pluginsImpl) RegisterExtensions() error {
	pi.Update()

	pi.lock.RLock()
	defer pi.lock.RUnlock()

	for _, p := range pi.plugins {
		if !p.IsValid() {
			continue
		}
		for _, m := range p.GetDescriptor().AccessMethods {
			name := m.Name
			if m.Version != "" {
				name = name + runtime.VersionSeparator + m.Version
			}
			pi.ctx.Logger(internal.TAG).Debug("registering access method",
				"plugin", p.Name(),
				"type", name)
			pi.ctx.AccessMethods().Register(name, access.NewType(name, p, &m))
		}

		for _, u := range p.GetDescriptor().Uploaders {
			for _, c := range u.Constraints {
				if c.ContextType != "" && c.RepositoryType != "" && c.MediaType != "" {
					hdlr, err := blob.New(p, u.Name, nil)
					if err != nil {
						pi.ctx.Logger(internal.TAG).Error("cannot create blob handler fpr plugin", "plugin", p.Name(), "handler", u.Name)
					} else {
						pi.ctx.Logger(internal.TAG).Debug("registering repository blob handler",
							"context", c.ContextType+":"+c.RepositoryType,
							"plugin", p.Name(),
							"handler", u.Name)
						pi.ctx.BlobHandlers().Register(hdlr, cpi.ForRepo(c.ContextType, c.RepositoryType), cpi.ForMimeType(c.MediaType))
					}
				}
			}
		}
	}
	return nil
}
