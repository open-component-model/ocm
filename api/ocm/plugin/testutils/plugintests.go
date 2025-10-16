package testutils

import (
	"github.com/mandelsoft/goutils/testutils"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugindirattr"
	"ocm.software/ocm/api/ocm/plugin/cache"
	"ocm.software/ocm/api/ocm/plugin/plugins"
)

type TempPluginDir = testutils.TempDir

func ConfigureTestPlugins2(ctx ocm.ContextProvider, path string) (TempPluginDir, plugins.Set, error) {
	t, err := ConfigureTestPlugins(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	return t, plugincacheattr.Get(ctx), nil
}

func ConfigureTestPlugins(ctx ocm.ContextProvider, path string) (TempPluginDir, error) {
	t, err := testutils.NewTempDir(testutils.WithDirContent(path))
	if err != nil {
		return nil, err
	}
	cache.DirectoryCache.Reset()
	plugindirattr.Set(ctx.OCMContext(), t.Path())
	return t, nil
}
