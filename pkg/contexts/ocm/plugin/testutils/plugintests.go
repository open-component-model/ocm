package testutils

import (
	"github.com/mandelsoft/goutils/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
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
