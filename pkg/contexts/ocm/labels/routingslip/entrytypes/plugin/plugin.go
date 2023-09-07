// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
)

type plug = plugin.Plugin

// PluginHandler is a shared object between the AccessMethod implementation and the Entry implementation. The
// object knows the actual plugin and can therefore forward the method calls to corresponding cli commands.
type PluginHandler struct {
	plug
}

func NewPluginHandler(p plugin.Plugin) *PluginHandler {
	return &PluginHandler{plug: p}
}

func (p *PluginHandler) Describe(spec *Entry, ctx cpi.Context) string {
	sspec := p.GetValueSetDescriptor(descriptor.PURPOSE_ROUTINGSLIP, spec.GetKind(), spec.GetVersion())
	if sspec == nil {
		return "unknown type " + spec.GetType()
	}
	info, err := p.Validate(spec)
	if err != nil {
		return err.Error()
	}
	return info.Short
}

func (p *PluginHandler) Validate(spec *Entry) (*ppi.ValueSetInfo, error) {
	data, err := spec.GetRaw()
	if err != nil {
		return nil, err
	}
	return p.plug.ValidateValueSet(descriptor.PURPOSE_ROUTINGSLIP, data)
}
