// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugincacheattr

import (
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
)

const (
	ATTR_KEY = "github.com/mandelsoft/ocm/plugins"
)

////////////////////////////////////////////////////////////////////////////////

func Get(ctx ocm.Context) cache.Cache {
	path := plugindirattr.Get(ctx)

	// avoid dead lock reading attribute during attribute creation
	return ctx.GetAttributes().GetOrCreateAttribute(ATTR_KEY, func(ctx datacontext.Context) interface{} {
		return cache.New(ctx.(ocm.Context), path)
	}).(cache.Cache)
}

func Set(ctx ocm.Context, cache cache.Cache) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, cache)
}
