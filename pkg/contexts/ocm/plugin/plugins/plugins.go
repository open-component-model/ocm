// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"encoding/json"
	"sync"

	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Set = *pluginsImpl

type pluginsImpl struct {
	lock sync.RWMutex

	updater cfgcpi.Updater
	ctx     cpi.Context
	base    cache.PluginDir
	configs map[string]json.RawMessage
	plugins map[string]plugin.Plugin
}

var _ config.Target = (*pluginsImpl)(nil)

func New(ctx cpi.Context, path string) Set {
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

func (pi *pluginsImpl) GetContext() cpi.Context {
	return pi.ctx
}

func (pi *pluginsImpl) Update() {
	err := pi.updater.Update()
	if err != nil {
		pi.ctx.Logger(descriptor.REALM).Error("config update failed", "error", err.Error())
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
