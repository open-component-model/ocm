// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
)

type Plugin = *pluginImpl

// //nolint: errname // is no error.
type pluginImpl struct {
	name       string
	descriptor *internal.Descriptor
	path       string
	error      string
}

func NewPlugin(name string, path string, desc *internal.Descriptor, errmsg string) Plugin {
	return &pluginImpl{
		name:       name,
		path:       path,
		descriptor: desc,
		error:      errmsg,
	}
}

func (p *pluginImpl) GetDescriptor() *internal.Descriptor {
	return p.descriptor
}

func (p *pluginImpl) Name() string {
	return p.name
}

func (p *pluginImpl) Path() string {
	return p.path
}

func (p *pluginImpl) Version() string {
	if !p.IsValid() {
		return "-"
	}
	return p.descriptor.PluginVersion
}

func (p *pluginImpl) IsValid() bool {
	return p.descriptor != nil
}

func (p *pluginImpl) Error() string {
	return p.error
}

func (p *pluginImpl) GetAccessMethodDescriptor(name, version string) *internal.AccessMethodDescriptor {
	if !p.IsValid() {
		return nil
	}

	var fallback internal.AccessMethodDescriptor
	fallbackFound := false
	for _, m := range p.descriptor.AccessMethods {
		if m.Name == name {
			if m.Version == version {
				return &m
			}
			if m.Version == "" || m.Version == "v1" {
				fallback = m
				fallbackFound = true
			}
		}
	}
	if fallbackFound && (version == "" || version == "v1") {
		return &fallback
	}
	return nil
}

func (p *pluginImpl) Message() string {
	if p.IsValid() {
		return p.descriptor.Short
	}
	if p.error != "" {
		return "Error: " + p.error
	}
	return "unknown state"
}
