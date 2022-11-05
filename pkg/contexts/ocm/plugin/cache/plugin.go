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
	mappings   *internal.Registry[*internal.UploaderDescriptor]
	uploaders  map[string]*internal.Registry[*internal.UploaderDescriptor]
}

func NewPlugin(name string, path string, desc *internal.Descriptor, errmsg string) Plugin {
	reg := internal.NewRegistry[*internal.UploaderDescriptor]()
	uploaders := map[string]*internal.Registry[*internal.UploaderDescriptor]{}

	for i := range desc.Uploaders {
		d := desc.Uploaders[i]
		nested := internal.NewRegistry[*internal.UploaderDescriptor]()
		for _, c := range d.Costraints {
			reg.Register(c, &d)
			nested.Register(c, &d)
		}
		uploaders[d.Name] = nested
	}
	return &pluginImpl{
		name:       name,
		path:       path,
		descriptor: desc,
		error:      errmsg,

		mappings:  reg,
		uploaders: uploaders,
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

func (p *pluginImpl) LookupUploader(name string, artType, mediaType string) *internal.UploaderDescriptor {
	if !p.IsValid() {
		return nil
	}

	if name == "" {
		if d, ok := p.mappings.LookupHandler(artType, mediaType); ok {
			return d
		}
	}

	u := p.uploaders[name]
	if u == nil {
		return nil
	}
	if d, ok := u.LookupHandler(artType, mediaType); ok {
		return d
	}
	return nil
}

func (p *pluginImpl) GetUploaderDescriptor(name string) *internal.UploaderDescriptor {
	if !p.IsValid() {
		return nil
	}

	for _, m := range p.descriptor.Uploaders {
		if m.Name == name {
			return &m
		}
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
