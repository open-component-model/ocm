// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type plugin struct {
	name       string
	version    string
	descriptor internal.Descriptor
	options    Options

	methods      map[string]AccessMethod
	accessScheme runtime.Scheme
}

func NewPlugin(name string, version string) Plugin {
	var rt runtime.VersionedTypedObject
	return &plugin{
		name:         name,
		version:      version,
		methods:      map[string]AccessMethod{},
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

func (p *plugin) RegisterAccessMethod(m AccessMethod) error {
	if p.GetAccessMethod(m.Name(), m.Version()) != nil {
		n := m.Name()
		if m.Version() != "" {
			n += "/" + m.Version()
		}
		return errors.ErrAlreadyExists(errors.KIND_ACCESSMETHOD, n)
	}

	vers := m.Version()
	if vers == "" {
		meth := internal.AccessMethodDescriptor{
			Name: m.Name(),
		}
		p.descriptor.AccessMethods = append(p.descriptor.AccessMethods, meth)
		p.accessScheme.RegisterByDecoder(m.Name(), m)
		vers = "v1"
		p.methods[m.Name()] = m
	}
	meth := internal.AccessMethodDescriptor{
		Name:    m.Name(),
		Version: vers,
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
