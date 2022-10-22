//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package clisupport

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/errors"
)

type ConfigProvider interface {
	CreateOptions() ConfigOptions
	GetConfigFor(opts ConfigOptions) (Config, error)
}

type ConfigTypeOptionSetConfigProvider interface {
	ConfigProvider
	ConfigOptionTypeSet
}

type typedConfigProvider struct {
	ConfigOptionTypeSet
}

var _ ConfigTypeOptionSetConfigProvider = (*typedConfigProvider)(nil)

func NewTypedConfigProvider(name string, desc string) ConfigTypeOptionSetConfigProvider {
	set := NewConfigOptionSet(name, NewValueMapOptionType(name, desc+" (YAML)"), NewStringOptionType(name+"Type", "type of "+desc))
	return &typedConfigProvider{
		ConfigOptionTypeSet: set,
	}
}

func (p *typedConfigProvider) GetConfigOptionTypeSet() ConfigOptionTypeSet {
	return p
}

func (p *typedConfigProvider) ApplyConfig(options ConfigOptions, config Config) error {
	cfg, err := p.GetConfigFor(options)
	if err != nil {
		return err
	}
	if cfg != nil {
		config[p.Name()] = cfg
	}
	return nil
}

func (p *typedConfigProvider) GetConfigFor(opts ConfigOptions) (Config, error) {
	typv, _ := opts.GetValue(p.Name() + "Type")
	cfgv, _ := opts.GetValue(p.Name())

	var data Config
	if cfgv != nil {
		data = cfgv.(Config)
	}
	typ := typv.(string)

	if typ == "" && data != nil && data["type"] != nil {
		t := data["type"]
		if t != nil {
			if s, ok := t.(string); ok {
				typ = s
			} else {
				return nil, fmt.Errorf("type field must be a string")
			}
		}
	}

	if opts.Changed() || typ != "" {
		if typ == "" {
			return nil, fmt.Errorf("type required for explicitly configured options")
		}
		if data == nil {
			data = Config{}
		}
		if typ != "" {
			data["type"] = typ
		}
		if err := p.applyConfigForType(typ, opts, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (p *typedConfigProvider) applyConfigForType(name string, opts ConfigOptions, config Config) error {
	set := p.GetTypeSet(name)
	if set == nil {
		return errors.ErrUnknown(p.Name())
	}

	err := opts.FilterBy(p.HasSharedOptionType).Check(set, p.Name()+" type "+name)
	if err != nil {
		return err
	}
	handler, ok := set.(ConfigHandler)
	if !ok {
		return fmt.Errorf("no config handler defined for %s type %s", p.Name(), name)
	}
	return handler.ApplyConfig(opts, config)
}
