// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package pluginhdlr

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
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
		result = append(result, &Object{cache.GetPlugin(n)})
	}
	return result, nil
}

func (h *TypeHandler) Get(elemspec utils.ElemSpec) ([]output.Object, error) {
	cache := plugincacheattr.Get(h.octx.Context())

	p := cache.GetPlugin(elemspec.String())
	if p == nil {
		return nil, errors.ErrNotFound(ppi.KIND_PLUGIN, elemspec.String())
	}
	return []output.Object{p}, nil
}
