package cache

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/info"
	"ocm.software/ocm/api/utils/filelock"
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

	return maputils.OrderedKeys(c.plugins)
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

func (c *pluginDirImpl) add(name string, desc *descriptor.Descriptor, path string, errmsg string, list *errors.ErrorList) {
	c.plugins[name] = NewPlugin(name, path, desc, errmsg)
	if path != "" {
		src, err := readPluginInstalltionInfo(filepath.Dir(path), filepath.Base(path))
		if err != nil && list != nil {
			list.Add(fmt.Errorf("%s: %s", name, err.Error()))
			return
		}

		c.plugins[name].info = src
	}
	if errmsg != "" && list != nil {
		list.Add(fmt.Errorf("%s: %s", name, errmsg))
	}
}

func (c *pluginDirImpl) scan(path string) error {
	fs := osfs.OsFs

	ok, err := vfs.Exists(fs, path)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	ok, err = vfs.IsDir(fs, path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("plugin path %q is no directory", path)
	}
	lockfile, err := filelock.MutexFor(path)
	if err != nil {
		return err
	}

	var finalize finalizer.Finalizer
	defer finalize.Finalize()

	DirectoryCache.numOfScans++
	entries, err := vfs.ReadDir(fs, path)
	if err != nil {
		return err
	}
	list := errors.ErrListf("scanning %q", path)
	for _, fi := range entries {
		if fi.Mode()&0o001 != 0 {
			loop := finalize.Nested()
			lock, err := lockfile.Lock()
			if err != nil {
				return err
			}
			loop.Close(lock)

			execpath := filepath.Join(path, fi.Name())
			desc, err := getCachedPluginInfo(path, fi.Name())

			errmsg := ""

			if err != nil {
				errmsg = err.Error()
			} else {
				if desc.PluginName != fi.Name() {
					errmsg = fmt.Sprintf("nmatching plugin name %q", desc.PluginName)
				}
			}
			c.add(fi.Name(), desc, execpath, errmsg, list)
			loop.Finalize()
		}
	}
	return list.Result()
}

func GetCachedPluginInfo(dir string, name string) (*descriptor.Descriptor, error) {
	l, err := filelock.LockDir(dir)
	if err != nil {
		return nil, err
	}
	defer l.Close()
	return getCachedPluginInfo(dir, name)
}

func getCachedPluginInfo(dir string, name string) (*descriptor.Descriptor, error) {
	src, err := readPluginInstalltionInfo(dir, name)
	if err != nil {
		return nil, err
	}
	execpath := filepath.Join(dir, name)
	if !src.IsValidPluginInfo(execpath) {
		mod, err := src.UpdatePluginInfo(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if mod {
			err := writePluginInstallationInfo(src, dir, name)
			if err != nil {
				return nil, err
			}
		}
	}
	return src.PluginInfo.Descriptor, nil
}

func GetPluginInfo(execpath string) (*descriptor.Descriptor, error) {
	result, err := Exec(execpath, nil, nil, nil, info.NAME)
	if err != nil {
		return nil, err
	}

	// TODO: Version handling by scheme
	var desc descriptor.Descriptor
	if err = json.Unmarshal(result, &desc); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal plugin descriptor: %s", err.Error())
	}
	return &desc, nil
}
