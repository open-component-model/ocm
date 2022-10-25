// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/info"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

var PKG = logging.Package()

type Cache = *cacheImpl

var _ config.Target = (*cacheImpl)(nil)

type cacheImpl struct {
	lock sync.RWMutex

	updater cfgcpi.Updater
	ctx     ocm.Context
	plugins map[string]plugin.Plugin
	configs map[string]json.RawMessage
}

func New(ctx ocm.Context, path string) Cache {
	c := &cacheImpl{
		ctx:     ctx,
		plugins: map[string]plugin.Plugin{},
		configs: map[string]json.RawMessage{},
	}
	c.updater = cfgcpi.NewUpdater(ctx.ConfigContext(), c)
	c.Update()
	if path != "" {
		c.scan(path)
	}
	return c
}

func (c *cacheImpl) Update() {
	err := c.updater.Update()
	if err != nil {
		c.ctx.Logger(PKG).Error("config update failed", "error", err)
	}
}

func (c *cacheImpl) PluginNames() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return utils.StringMapKeys(c.plugins)
}

func (c *cacheImpl) GetPlugin(name string) plugin.Plugin {
	c.Update()
	c.lock.RLock()
	defer c.lock.RUnlock()

	p, ok := c.plugins[name]
	if ok {
		return p
	}
	return nil
}

func (c *cacheImpl) add(name string, desc *internal.Descriptor, path string, errmsg string, list *errors.ErrorList) {
	c.plugins[name] = plugin.NewPlugin(name, path, c.configs[name], desc, errmsg)
	if errmsg != "" && list != nil {
		list.Add(fmt.Errorf("%s: %s", name, errmsg))
	}
}

func (c *cacheImpl) scan(path string) error {
	fs := osfs.New()
	entries, err := vfs.ReadDir(fs, path)
	if err != nil {
		return err
	}
	list := errors.ErrListf("scanning %q", path)
	for _, fi := range entries {
		if fi.Mode()&0o001 != 0 {
			execpath := filepath.Join(path, fi.Name())
			config := c.configs[fi.Name()]
			result, err := plugin.Exec(execpath, config, nil, nil, info.NAME)
			if err != nil {
				c.add(fi.Name(), nil, execpath, err.Error(), list)
				continue
			}

			// TODO: Version handling by scheme
			var desc internal.Descriptor
			if err = json.Unmarshal(result, &desc); err != nil {
				c.add(fi.Name(), nil, execpath, fmt.Sprintf("cannot unmarshal plugin descriptor: %s", err.Error()), list)
				continue
			}

			if desc.PluginName != fi.Name() {
				c.add(fi.Name(), nil, execpath, fmt.Sprintf("nmatching plugin name %q", desc.PluginName), list)
				continue
			}
			c.add(desc.PluginName, &desc, execpath, "", nil)
		}
	}
	return list.Result()
}

func (c *cacheImpl) ConfigurePlugin(name string, config json.RawMessage) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.configs[name] = config
	if p := c.plugins[name]; p != nil {
		p.SetConfig(config)
	}
}

// RegisterExtensions registers all the extension provided the found plugin
// at the given context. If no context is given, the cache context is used.
func (c *cacheImpl) RegisterExtensions(ctx ocm.Context) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if ctx == nil {
		ctx = c.ctx
	}
	for _, p := range c.plugins {
		if !p.IsValid() {
			continue
		}
		for _, m := range p.GetDescriptor().AccessMethods {
			name := m.Name
			if m.Version != "" {
				name = name + runtime.VersionSeparator + m.Version
			}
			ctx.AccessMethods().Register(name, access.NewType(name, m.Long, p))
		}
	}
	return nil
}
