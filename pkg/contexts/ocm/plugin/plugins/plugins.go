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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/config"
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
	c := &pluginsImpl{
		ctx:     ctx,
		configs: map[string]json.RawMessage{},
		plugins: map[string]plugin.Plugin{},
	}
	c.updater = cfgcpi.NewUpdater(ctx.ConfigContext(), c)
	c.Update()
	c.base = cache.Get(path)
	for _, n := range c.base.PluginNames() {
		c.plugins[n] = plugin.NewPlugin(c.base.Get(n), c.configs[n])
	}
	return c
}

func (c *pluginsImpl) Update() {
	err := c.updater.Update()
	if err != nil {
		c.ctx.Logger(plugin.PKG).Error("config update failed", "error", err)
	}
}

func (c *pluginsImpl) ConfigurePlugin(name string, config json.RawMessage) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.configs[name] = config
}

func (c *pluginsImpl) PluginNames() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return utils.StringMapKeys(c.plugins)
}

func (c *pluginsImpl) Get(name string) plugin.Plugin {
	c.lock.RLock()
	defer c.lock.RUnlock()

	p, ok := c.plugins[name]
	if ok {
		return p
	}
	return nil
}

// RegisterExtensions registers all the extension provided the found plugin
// at the given context. If no context is given, the cache context is used.
func (c *pluginsImpl) RegisterExtensions() error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, p := range c.plugins {
		if !p.IsValid() {
			continue
		}
		for _, m := range p.GetDescriptor().AccessMethods {
			name := m.Name
			if m.Version != "" {
				name = name + runtime.VersionSeparator + m.Version
			}
			c.ctx.AccessMethods().Register(name, access.NewType(name, p, m.Description, m.Format))
		}
	}
	return nil
}
