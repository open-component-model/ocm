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

package cpi

// This is the Context Provider Interface for credential providers

import (
	"github.com/gardener/ocm/pkg/config/core"
	"github.com/gardener/ocm/pkg/runtime"
)

const KIND_CONFIGTYPE = core.KIND_CONFIGTYPE

const CONTEXT_TYPE = core.CONTEXT_TYPE

type Context = core.Context
type Config = core.Config
type ConfigType = core.ConfigType
type ConfigTypeScheme = core.ConfigTypeScheme
type GenericConfig = core.GenericConfig

var DefaultContext = core.DefaultContext

func RegisterConfigType(name string, atype ConfigType) {
	core.DefaultConfigTypeScheme.Register(name, atype)
}

func NewGenericConfig(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	return core.NewGenericConfig(data, unmarshaler)
}

func ToGenericConfig(c Config) (*GenericConfig, error) {
	return core.ToGenericConfig(c)
}

func NewConfigTypeScheme() ConfigTypeScheme {
	return core.NewConfigTypeScheme(nil)
}

func IsGeneric(cfg Config) bool {
	return core.IsGeneric(cfg)
}

////////////////////////////////////////////////////////////////////////////////

type Updater = core.Updater

func NewUpdate(ctx Context) Updater {
	return core.NewUpdater(ctx)
}

////////////////////////////////////////////////////////////////////////////////

func ErrNoContext(name string) error {
	return core.ErrNoContext(name)
}

func IsErrNoContext(err error) bool {
	return core.IsErrNoContext(err)
}

func IsErrConfigNotApplicable(err error) bool {
	return core.IsErrConfigNotApplicable(err)
}
