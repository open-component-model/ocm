// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package pluginhdlr

import (
	"strings"

	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/v2/pkg/errors"
)

func Elem(e interface{}) plugin.Plugin {
	return e.(*Object).Plugin
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	Plugin plugin.Plugin
}

type Manifest struct {
	Element *plugin.Descriptor `json:"element"`
}

func (o *Object) AsManifest() interface{} {
	return &Manifest{
		o.Plugin.GetDescriptor(),
	}
}

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	octx clictx.OCM
}

func NewTypeHandler(octx clictx.OCM) utils.TypeHandler {
	return &TypeHandler{
		octx: octx,
	}
}

func (h *TypeHandler) Close() error {
	return nil
}

func (h *TypeHandler) All() ([]output.Object, error) {
	cache := plugincacheattr.Get(h.octx.Context())
	result := []output.Object{}

	for _, n := range cache.PluginNames() {
		result = append(result, &Object{cache.Get(n)})
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	cache := plugincacheattr.Get(h.octx.Context())

	p := cache.Get(elemspec.String())
	if p == nil {
		objs := Lookup(elemspec.String(), cache)
		if len(objs) == 0 {
			return nil, errors.ErrNotFound(descriptor.KIND_PLUGIN, elemspec.String())
		}
		return objs, nil
	}
	return []output.Object{&Object{p}}, nil
}

func Lookup(prefix string, cache plugins.Set) []output.Object {
	var objs []output.Object
	prefix = prefix + "."
	for _, n := range cache.PluginNames() {
		if strings.HasPrefix(n, prefix) {
			objs = append(objs, &Object{cache.Get(n)})
		}
	}
	return objs
}
