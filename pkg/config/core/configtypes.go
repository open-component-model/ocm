// Copyright 2022 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"reflect"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/runtime"
	"github.com/modern-go/reflect2"
)

type ConfigType interface {
	runtime.TypedObjectDecoder
	runtime.VersionedTypedObject
}

type ConfigTypeScheme interface {
	runtime.Scheme

	GetConfigType(name string) ConfigType
	Register(name string, atype ConfigType)

	DecodeConfig(data []byte, unmarshaler runtime.Unmarshaler) (Config, error)
	CreateConfig(obj runtime.TypedObject) (Config, error)
}

type configTypeScheme struct {
	runtime.Scheme
}

func NewConfigTypeScheme(defaultRepoDecoder runtime.TypedObjectDecoder) ConfigTypeScheme {
	var rt Config
	scheme := runtime.MustNewDefaultScheme(&rt, &GenericConfig{}, true, defaultRepoDecoder)
	return &configTypeScheme{scheme}
}

func (t *configTypeScheme) AddKnowntypes(s ConfigTypeScheme) {
	t.Scheme.AddKnownTypes(s)
}

func (t *configTypeScheme) GetConfigType(name string) ConfigType {
	d := t.GetDecoder(name)
	if d == nil {
		return nil
	}
	return d.(ConfigType)
}

func (t *configTypeScheme) RegisterByDecoder(name string, decoder runtime.TypedObjectDecoder) error {
	if _, ok := decoder.(ConfigType); !ok {
		errors.ErrInvalid("type", reflect.TypeOf(decoder).String())
	}
	return t.Scheme.RegisterByDecoder(name, decoder)
}

func (t *configTypeScheme) AddKnownTypes(scheme runtime.Scheme) {
	if _, ok := scheme.(ConfigTypeScheme); !ok {
		panic("can only add RepositoryTypeSchemes")
	}
	t.Scheme.AddKnownTypes(scheme)
}

func (t *configTypeScheme) Register(name string, rtype ConfigType) {
	t.Scheme.RegisterByDecoder(name, rtype)
}

func (t *configTypeScheme) DecodeConfig(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	obj, err := t.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	if spec, ok := obj.(Config); ok {
		return spec, nil
	}
	return nil, fmt.Errorf("invalid object type: yield %T instead of Config", obj)
}

func (t *configTypeScheme) CreateConfig(obj runtime.TypedObject) (Config, error) {
	if s, ok := obj.(Config); ok {
		return s, nil
	}
	if u, ok := obj.(*runtime.UnstructuredTypedObject); ok {
		raw, err := u.GetRaw()
		if err != nil {
			return nil, err
		}
		return t.DecodeConfig(raw, runtime.DefaultJSONEncoding)
	}
	return nil, fmt.Errorf("invalid object type %T for repository specs", obj)
}

// DefaultConfigTypeScheme contains all globally known access serializer
var DefaultConfigTypeScheme = NewConfigTypeScheme(nil)

////////////////////////////////////////////////////////////////////////////////

type Evaluator interface {
	Evaluate(ctx Context) (Config, error)
}

type GenericConfig struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

func IsGeneric(cfg Config) bool {
	_, ok := cfg.(*GenericConfig)
	return ok
}

func NewGenericConfig(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	unstr := &runtime.UnstructuredVersionedTypedObject{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, unstr)
	if err != nil {
		return nil, err
	}
	return &GenericConfig{*unstr}, nil
}

func ToGenericConfig(c Config) (*GenericConfig, error) {
	if reflect2.IsNil(c) {
		return nil, nil
	}
	if g, ok := c.(*GenericConfig); ok {
		return g, nil
	}
	u, err := runtime.ToUnstructuredVersionedTypedObject(c)
	if err != nil {
		return nil, err
	}
	return &GenericConfig{*u}, nil
}

func (s *GenericConfig) Evaluate(ctx Context) (Config, error) {
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	cfg, err := ctx.ConfigTypes().DecodeConfig(raw, runtime.DefaultJSONEncoding)
	if IsGeneric(cfg) {
		return nil, errors.ErrUnknown(KIND_CONFIGTYPE, s.GetType())
	}
	return cfg, nil
}

func (s *GenericConfig) ApplyTo(ctx Context, target interface{}) error {
	spec, err := s.Evaluate(ctx)
	if err != nil {
		return err
	}
	return spec.ApplyTo(ctx, target)
}

var _ Config = &GenericConfig{}

////////////////////////////////////////////////////////////////////////////////
