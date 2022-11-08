// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugincacheattr

import (
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	ATTR_KEY = "github.com/mandelsoft/ocm/plugins"
)

////////////////////////////////////////////////////////////////////////////////

func Get(ctx ocm.Context) plugins.Set {
	path := plugindirattr.Get(ctx)

	// avoid dead lock reading attribute during attribute creation
	return ctx.GetAttributes().GetOrCreateAttribute(ATTR_KEY, func(ctx datacontext.Context) interface{} {
		return plugins.New(ctx.(ocm.Context), path)
	}).(plugins.Set)
}

func Set(ctx ocm.Context, cache cache.PluginDir) error {
	return ctx.GetAttributes().SetAttribute(ATTR_KEY, cache)
}

func RegisterBlobHandler(ctx ocm.Context, pname, name string, artType, mediaType string, target json.RawMessage) error {
	set := Get(ctx)
	if set == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}
	return set.RegisterBlobHandler(pname, name, artType, mediaType, target)
}

func RegisterDownloadHandler(ctx ocm.Context, pname, name string, artType, mediaType string) error {
	set := Get(ctx)
	if set == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}
	return set.RegisterDownloadHandler(pname, name, artType, mediaType)
}
