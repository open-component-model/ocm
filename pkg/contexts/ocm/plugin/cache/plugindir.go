// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/info"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

type PluginDir = *pluginDirImpl

type pluginDirImpl struct {
	lock sync.RWMutex

	plugins map[string]Plugin
}

func NewDir(path string) PluginDir {
	c := &pluginDirImpl{
		plugins: map[string]Plugin{},
	}
	if path != "" {
		c.scan(path)
	}
	return c
}

func (c *pluginDirImpl) PluginNames() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return utils.StringMapKeys(c.plugins)
}

func (c *pluginDirImpl) Get(name string) Plugin {
	c.lock.RLock()
	defer c.lock.RUnlock()

	p, ok := c.plugins[name]
	if ok {
		return p
	}
	return nil
}

func (c *pluginDirImpl) add(name string, desc *internal.Descriptor, path string, errmsg string, list *errors.ErrorList) {
	c.plugins[name] = NewPlugin(name, path, desc, errmsg)
	if errmsg != "" && list != nil {
		list.Add(fmt.Errorf("%s: %s", name, errmsg))
	}
}

func (c *pluginDirImpl) scan(path string) error {
	DirectoryCache.numOfScans++
	fs := osfs.New()
	entries, err := vfs.ReadDir(fs, path)
	if err != nil {
		return err
	}
	list := errors.ErrListf("scanning %q", path)
	for _, fi := range entries {
		if fi.Mode()&0o001 != 0 {
			execpath := filepath.Join(path, fi.Name())
			result, err := Exec(execpath, nil, nil, nil, info.NAME)
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
