// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package core

import (
	"context"
	"reflect"

	"github.com/gardener/ocm/pkg/datacontext"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/runtime"
)

type ConfigSelector interface {
	Select(Config) bool
}
type ConfigSelectorFunction func(Config) bool

func (f ConfigSelectorFunction) Select(cfg Config) bool { return f(cfg) }

var AllConfigs = AppliedConfigSelectorFunction(func(*AppliedConfig) bool { return true })

const AllGenerations int64 = 0

const CONTEXT_TYPE = "config.context.gardener.cloud"

type Context interface {
	datacontext.Context

	AttributesContext() datacontext.AttributesContext

	// Info provides the context for nested configuration evaluation
	Info() string
	// WithInfo provides the same context withh additional nesting info
	WithInfo(desc string) Context

	ConfigTypes() ConfigTypeScheme

	GetConfigForData(data []byte, unmarshaler runtime.Unmarshaler) (Config, error)

	// ApplyData applies the config given by a byte stream to the config store
	// If the config type is not known, a generic config is stored and returned.
	// In this case an unknown error for kind KIND_CONFIGTYPE is returned.
	ApplyData(data []byte, unmarshaler runtime.Unmarshaler, desc string) (Config, error)
	// ApplyConfig applies the config to the config store
	ApplyConfig(spec Config, desc string) error

	GetConfigForType(generation int64, typ string) (int64, []Config)
	GetConfigForName(generation int64, name string) (int64, []Config)
	GetConfig(generation int64, selector ConfigSelector) (int64, []Config)

	Generation() int64
	ApplyTo(gen int64, target interface{}) (int64, error)
}

var key = reflect.TypeOf(_context{})

// DefaultContext is the default context initialized by init functions
var DefaultContext = Builder{}.New()

// ForContext returns the Context to use for context.Context.
// This is eiter an explicit context or the default context.
// The returned context incorporates the given context.
func ForContext(ctx context.Context) Context {
	return datacontext.ForContextByKey(ctx, key, DefaultContext).(Context)
}

////////////////////////////////////////////////////////////////////////////////

type __context struct {
	datacontext.Context

	sharedAttributes datacontext.AttributesContext

	knownConfigTypes ConfigTypeScheme

	configs *ConfigStore
}

type _context struct {
	*__context
	description string
}

var _ Context = &_context{}

func newContext(shared datacontext.AttributesContext, reposcheme ConfigTypeScheme) Context {
	c := &_context{
		__context: &__context{
			sharedAttributes: shared,
			knownConfigTypes: reposcheme,
			configs:          NewConfigStore(),
		},
	}
	c.Context = datacontext.NewContextBase(c, CONTEXT_TYPE, key, shared.GetAttributes())
	return c
}

func (c *_context) Info() string {
	return c.description
}

func (c *_context) WithInfo(desc string) Context {
	if c.description != "" {
		desc = desc + "--" + c.description
	}
	return &_context{c.__context, desc}
}

func (c *_context) AttributesContext() datacontext.AttributesContext {
	return c.sharedAttributes
}

func (c *_context) ConfigTypes() ConfigTypeScheme {
	return c.knownConfigTypes
}

func (c *_context) ConfigForData(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	return c.knownConfigTypes.DecodeConfig(data, unmarshaler)
}

func (c *_context) GetConfigForData(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	spec, err := c.knownConfigTypes.DecodeConfig(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func (c *_context) ApplyConfig(spec Config, desc string) error {
	var unknown error
	spec = (&AppliedConfig{config: spec}).eval(c)
	if IsGeneric(spec) {
		unknown = errors.ErrUnknown(KIND_CONFIGTYPE, spec.GetType())
	}

	ctx := c.WithInfo(desc)
	err := spec.ApplyTo(c, ctx)
	if IsErrNoContext(err) {
		err = unknown
	}
	err = errors.Wrapf(err, ctx.Info())
	c.configs.Apply(spec, ctx.Info())
	return err
}

func (c *_context) joinInfo(desc string) string {
	if c.description == "" {
		return desc
	}
	return desc + "--" + c.description
}

func (c *_context) ApplyData(data []byte, unmarshaler runtime.Unmarshaler, desc string) (Config, error) {
	spec, err := c.knownConfigTypes.DecodeConfig(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return spec, c.ApplyConfig(spec, desc)
}

func (c *_context) selector(gen int64, selector ConfigSelector) AppliedConfigSelector {
	if gen <= 0 {
		return AppliedConfigSelectorFor(selector)
	}
	if selector == nil {
		return AppliedGenerationSelector(gen)
	}
	return AppliedAndSelector(AppliedGenerationSelector(gen), AppliedConfigSelectorFor(selector))
}

func (c *_context) Generation() int64 {
	return c.configs.Generation()
}

func (c *_context) ApplyTo(gen int64, target interface{}) (int64, error) {
	cur := c.configs.Generation()
	if cur <= gen {
		return gen, nil
	}
	cur, cfgs := c.configs.GetConfigForSelector(c, AppliedGenerationSelector(gen))

	list := errors.ErrListf("config apply errors")
	for _, cfg := range cfgs {
		list.Add(errors.Wrapf(cfg.config.ApplyTo(c, target), cfg.description))
	}
	return cur, list.Result()
}

func (c *_context) GetConfig(gen int64, selector ConfigSelector) (int64, []Config) {
	gen, cfgs := c.configs.GetConfigForSelector(c, c.selector(gen, selector))
	return gen, cfgs.Configs()
}

func (c *_context) GetConfigForName(gen int64, name string) (int64, []Config) {
	gen, cfgs := c.configs.GetConfigForName(c, name, c.selector(gen, nil))
	return gen, cfgs.Configs()
}

func (c *_context) GetConfigForType(gen int64, typ string) (int64, []Config) {
	gen, cfgs := c.configs.GetConfigForType(c, typ, c.selector(gen, nil))
	return gen, cfgs.Configs()
}
