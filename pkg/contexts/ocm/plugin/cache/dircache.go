// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"sync"
)

type PluginDirCache = *pluginDirCache

type pluginDirCache struct {
	lock          sync.Mutex
	directories   map[string]PluginDir
	numOfScans    int
	numOfRequests int
}

var DirectoryCache = &pluginDirCache{
	directories: map[string]PluginDir{},
}

func (c *pluginDirCache) Count() int {
	return c.numOfScans
}

func (c *pluginDirCache) Requests() int {
	return c.numOfRequests
}

func (c *pluginDirCache) Reset() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.numOfScans = 0
	c.numOfRequests = 0
	c.directories = map[string]PluginDir{}
}

func (c *pluginDirCache) Get(path string) PluginDir {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.numOfRequests++
	found := c.directories[path]
	if found == nil {
		found = NewDir(path)
		c.directories[path] = found
	}
	return found
}

func Get(path string) PluginDir {
	return DirectoryCache.Get(path)
}
