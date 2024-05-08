package plugincacheattr

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
)

const (
	ATTR_KEY = "github.com/mandelsoft/ocm/plugins"
)

////////////////////////////////////////////////////////////////////////////////

func Get(ctxp ocm.ContextProvider) plugins.Set {
	ctx := ctxp.OCMContext()
	path := plugindirattr.Get(ctx)

	// avoid dead lock reading attribute during attribute creation
	return ctx.GetAttributes().GetOrCreateAttribute(ATTR_KEY, func(ctx datacontext.Context) interface{} {
		return plugins.New(ctx.(ocm.Context), path)
	}).(plugins.Set)
}

func Set(ctx ocm.Context, cache cache.PluginDir) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, cache)
}
