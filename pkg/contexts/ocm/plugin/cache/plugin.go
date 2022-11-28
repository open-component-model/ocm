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
	name        string
	source      *PluginSource
	descriptor  *internal.Descriptor
	path        string
	error       string
	uploaders   *ConstraintRegistry[internal.UploaderDescriptor, internal.UploaderKey]
	downloaders *ConstraintRegistry[internal.DownloaderDescriptor, internal.DownloaderKey]
}

func NewPlugin(name string, path string, desc *internal.Descriptor, errmsg string) Plugin {
	p := &pluginImpl{
		name:       name,
		path:       path,
		descriptor: desc,
		error:      errmsg,
	}
	if desc != nil {
		p.uploaders = NewConstraintRegistry[internal.UploaderDescriptor, internal.UploaderKey](desc.Uploaders)
		p.downloaders = NewConstraintRegistry[internal.DownloaderDescriptor, internal.DownloaderKey](desc.Downloaders)
	} else {
		p.uploaders = NewConstraintRegistry[internal.UploaderDescriptor, internal.UploaderKey](nil)
		p.downloaders = NewConstraintRegistry[internal.DownloaderDescriptor, internal.DownloaderKey](nil)
	}
	return p
}

func (p *pluginImpl) GetSource() *PluginSource {
	return p.source
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

func (p *pluginImpl) LookupDownloader(name string, artType, mediaType string) internal.List[*internal.DownloaderDescriptor] {
	if !p.IsValid() {
		return nil
	}

	return p.downloaders.LookupFor(name, internal.NewDownloaderKey(artType, mediaType))
}

func (p *pluginImpl) GetDownloaderDescriptor(name string) *internal.DownloaderDescriptor {
	if !p.IsValid() {
		return nil
	}
	return p.descriptor.Downloaders.Get(name)
}

func (p *pluginImpl) LookupUploader(name string, artType, mediaType string) internal.List[*internal.UploaderDescriptor] {
	if !p.IsValid() {
		return nil
	}

	return p.uploaders.LookupFor(name, internal.UploaderKey{}.SetArtifact(artType, mediaType))
}

func (p *pluginImpl) LookupUploadersForKeys(name string, keys internal.UploaderKeySet) internal.List[*internal.UploaderDescriptor] {
	if !p.IsValid() {
		return nil
	}

	var r internal.List[*internal.UploaderDescriptor]
	for k := range keys {
		r = r.MergeWith(p.uploaders.LookupFor(name, k))
	}
	return r
}

func (p *pluginImpl) LookupUploaderKeys(name string, artType, mediaType string) internal.UploaderKeySet {
	if !p.IsValid() {
		return nil
	}

	return p.uploaders.LookupKeysFor(name, internal.UploaderKey{}.SetArtifact(artType, mediaType))
}

func (p *pluginImpl) GetUploaderDescriptor(name string) *internal.UploaderDescriptor {
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
