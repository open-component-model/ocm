// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type plugin struct {
	name       string
	version    string
	descriptor internal.Descriptor
	options    Options

	uploaders map[string]Uploader
	mappings  *internal.Registry[Uploader]

	methods      map[string]AccessMethod
	accessScheme runtime.Scheme
}

func NewPlugin(name string, version string) Plugin {
	var rt runtime.VersionedTypedObject
	return &plugin{
		name:         name,
		version:      version,
		methods:      map[string]AccessMethod{},
		uploaders:    map[string]Uploader{},
		mappings:     internal.NewRegistry[Uploader](),
		accessScheme: runtime.MustNewDefaultScheme(&rt, &runtime.UnstructuredVersionedTypedObject{}, false, nil),
		descriptor: internal.Descriptor{
			Version:       internal.VERSION,
			PluginName:    name,
			PluginVersion: version,
		},
	}
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Version() string {
	return p.version
}

func (p *plugin) Descriptor() internal.Descriptor {
	return p.descriptor
}

func (p *plugin) Options() *Options {
	return &p.options
}

func (p *plugin) SetLong(s string) {
	p.descriptor.Long = s
}

func (p *plugin) SetShort(s string) {
	p.descriptor.Short = s
}

func (p *plugin) RegisterUploader(arttype, mediatype string, u Uploader) error {
	old := p.uploaders[u.Name()]
	if old != nil && old != u {
		return fmt.Errorf("uploader name %q already in use", u.Name())
	}

	var d *UploaderDescriptor
	if old == nil {
		d = &UploaderDescriptor{
			Name:        u.Name(),
			Description: u.Description(),
			Costraints:  []UploaderKey{},
		}
		p.descriptor.Uploaders = append(p.descriptor.Uploaders, *d)
		d = &p.descriptor.Uploaders[len(p.descriptor.Uploaders)-1]
	} else {
		for i := range p.descriptor.Uploaders {
			if p.descriptor.Uploaders[i].Name == u.Name() {
				d = &p.descriptor.Uploaders[i]
			}
		}
	}
	p.uploaders[u.Name()] = u

	key := UploaderKey{
		ArtifactType: arttype,
		MediaType:    mediatype,
	}
	cur := p.mappings.GetHandler(key)
	if cur != nil && cur != u {
		return fmt.Errorf("uploader mapping key %q already in use", key)
	}
	if cur == nil {
		p.mappings.Register(key, u)

		if key.ArtifactType == "" {
			key.ArtifactType = "*"
		}
		if key.MediaType == "" {
			key.MediaType = "*"
		}
		d.Costraints = append(d.Costraints, key)
	}
	return nil
}

func (p *plugin) GetUploader(arttype, mediatype string) Uploader {
	u, _ := p.mappings.LookupHandler(arttype, mediatype)
	return u
}

func (p *plugin) RegisterAccessMethod(m AccessMethod) error {
	if p.GetAccessMethod(m.Name(), m.Version()) != nil {
		n := m.Name()
		if m.Version() != "" {
			n += runtime.VersionSeparator + m.Version()
		}
		return errors.ErrAlreadyExists(errors.KIND_ACCESSMETHOD, n)
	}

	vers := m.Version()
	if vers == "" {
		meth := internal.AccessMethodDescriptor{
			Name:        m.Name(),
			Description: m.Description(),
			Format:      m.Format(),
		}
		p.descriptor.AccessMethods = append(p.descriptor.AccessMethods, meth)
		p.accessScheme.RegisterByDecoder(m.Name(), m)
		vers = "v1"
		p.methods[m.Name()] = m
	}
	meth := internal.AccessMethodDescriptor{
		Name:        m.Name(),
		Version:     vers,
		Description: m.Description(),
		Format:      m.Format(),
	}
	p.descriptor.AccessMethods = append(p.descriptor.AccessMethods, meth)
	p.accessScheme.RegisterByDecoder(m.Name()+"/"+vers, m)
	p.methods[m.Name()+"/"+vers] = m
	return nil
}

func (p *plugin) DecodeAccessSpecification(data []byte) (AccessSpec, error) {
	o, err := p.accessScheme.Decode(data, nil)
	if err != nil {
		return nil, err
	}
	return o.(AccessSpec), nil
}

func (p *plugin) GetAccessMethod(name string, version string) AccessMethod {
	n := name
	if version != "" {
		n += "/" + version
	}
	return p.methods[n]
}
