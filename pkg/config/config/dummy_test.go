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

package config_test

import (
	"reflect"

	"github.com/gardener/ocm/pkg/config"
	"github.com/gardener/ocm/pkg/config/cpi"
	"github.com/gardener/ocm/pkg/runtime"
)

const (
	DummyType   = "Dummy"
	DummyTypeV1 = DummyType + "/v1"
)

func RegisterAt(reg cpi.ConfigTypeScheme) {
	reg.Register(DummyType, cpi.NewConfigType(DummyType, &Config{}))
	reg.Register(DummyType, cpi.NewConfigType(DummyTypeV1, &Config{}))
}

// Config describes a a dummy config
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Alice                       string `json:"alice,omitempty"`
	Bob                         string `json:"bob,omitempty"`
}

// NewConfig creates a new memory Config
func NewConfig(a, b string) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(DummyType),
		Alice:               a,
		Bob:                 b,
	}
}

func (a *Config) GetType() string {
	return DummyType
}

func (a *Config) ApplyTo(ctx config.Context, target interface{}) error {
	d, ok := target.(*dummyContext)
	if ok {
		d.applied = append(d.applied, a)
		return nil
	}
	return cpi.ErrNoContext(DummyType)
}

////////////////////////////////////////////////////////////////////////////////

var key = reflect.TypeOf(&dummyContext{})

func newDummy(ctx config.Context) *dummyContext {
	d := &dummyContext{
		config: ctx,
	}
	d.update()
	return d
}

type dummyContext struct {
	config         config.Context
	lastGeneration int64
	applied        []*Config
}

func (d *dummyContext) getApplied() []*Config {
	d.update()
	return d.applied
}
func (d *dummyContext) update() error {
	gen, err := d.config.ApplyTo(d.lastGeneration, d)
	d.lastGeneration = gen
	return err
}
