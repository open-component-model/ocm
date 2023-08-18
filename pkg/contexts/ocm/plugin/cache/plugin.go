// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
)

type Plugin = *pluginImpl

// //nolint: errname // is no error.
type pluginImpl struct {
	name        string
	source      *PluginSource
	descriptor  *descriptor.Descriptor
	path        string
	error       string
	uploaders   *ConstraintRegistry[descriptor.UploaderDescriptor, descriptor.UploaderKey]
	downloaders *ConstraintRegistry[descriptor.DownloaderDescriptor, descriptor.DownloaderKey]
}

func NewPlugin(name string, path string, desc *descriptor.Descriptor, errmsg string) Plugin {
	p := &pluginImpl{
		name:       name,
		path:       path,
		descriptor: desc,
		error:      errmsg,
	}
	if desc != nil {
		p.uploaders = NewConstraintRegistry[descriptor.UploaderDescriptor, descriptor.UploaderKey](desc.Uploaders)
		p.downloaders = NewConstraintRegistry[descriptor.DownloaderDescriptor, descriptor.DownloaderKey](desc.Downloaders)
	} else {
		p.uploaders = NewConstraintRegistry[descriptor.UploaderDescriptor, descriptor.UploaderKey](nil)
		p.downloaders = NewConstraintRegistry[descriptor.DownloaderDescriptor, descriptor.DownloaderKey](nil)
	}
	return p
}

func (p *pluginImpl) GetSource() *PluginSource {
	return p.source
}

func (p *pluginImpl) GetDescriptor() *descriptor.Descriptor {
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

func (p *pluginImpl) GetActionDescriptor(name string) *descriptor.ActionDescriptor {
	if !p.IsValid() {
		return nil
	}

	for _, a := range p.descriptor.Actions {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

func (p *pluginImpl) GetValueMergeHandlerDescriptor(name string) *descriptor.ValueMergeHandlerDescriptor {
	if !p.IsValid() {
		return nil
	}

	for _, a := range p.descriptor.ValueMergeHandlers {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

func (p *pluginImpl) GetValueMappingDescriptor(name string) *descriptor.ValueMergeHandlerDescriptor {
	if !p.IsValid() {
		return nil
	}

	for _, a := range p.descriptor.ValueMergeHandlers {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

func (p *pluginImpl) GetAccessMethodDescriptor(name, version string) *descriptor.AccessMethodDescriptor {
	if !p.IsValid() {
		return nil
	}

	var fallback descriptor.AccessMethodDescriptor
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

func (p *pluginImpl) LookupDownloader(name string, artType, mediaType string) descriptor.List[*descriptor.DownloaderDescriptor] {
	if !p.IsValid() {
		return nil
	}

	return p.downloaders.LookupFor(name, descriptor.NewDownloaderKey(artType, mediaType))
}

func (p *pluginImpl) GetDownloaderDescriptor(name string) *descriptor.DownloaderDescriptor {
	if !p.IsValid() {
		return nil
	}
	return p.descriptor.Downloaders.Get(name)
}

func (p *pluginImpl) LookupUploader(name string, artType, mediaType string) descriptor.List[*descriptor.UploaderDescriptor] {
	if !p.IsValid() {
		return nil
	}

	return p.uploaders.LookupFor(name, descriptor.UploaderKey{}.SetArtifact(artType, mediaType))
}

func (p *pluginImpl) LookupUploadersForKeys(name string, keys descriptor.UploaderKeySet) descriptor.List[*descriptor.UploaderDescriptor] {
	if !p.IsValid() {
		return nil
	}

	var r descriptor.List[*descriptor.UploaderDescriptor]
	for k := range keys {
		r = r.MergeWith(p.uploaders.LookupFor(name, k))
	}
	return r
}

func (p *pluginImpl) LookupUploaderKeys(name string, artType, mediaType string) descriptor.UploaderKeySet {
	if !p.IsValid() {
		return nil
	}

	return p.uploaders.LookupKeysFor(name, descriptor.UploaderKey{}.SetArtifact(artType, mediaType))
}

func (p *pluginImpl) GetUploaderDescriptor(name string) *descriptor.UploaderDescriptor {
	if !p.IsValid() {
		return nil
	}
	return p.descriptor.Uploaders.Get(name)
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
