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

package config

import (
	"context"

	core2 "github.com/open-component-model/ocm/pkg/contexts/config/core"
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const KIND_CONFIGTYPE = core2.KIND_CONFIGTYPE

var AllConfigs = core2.AllConfigs

const AllGenerations = core2.AllGenerations

const CONTEXT_TYPE = core2.CONTEXT_TYPE

type Context = core2.Context
type Config = core2.Config
type ConfigType = core2.ConfigType
type ConfigTypeScheme = core2.ConfigTypeScheme
type GenericConfig = core2.GenericConfig
type ConfigSelector = core2.ConfigSelector
type ConfigSelectorFunction = core2.ConfigSelectorFunction

func DefaultContext() core2.Context {
	return core2.DefaultContext
}
func ForContext(ctx context.Context) Context {
	return core2.ForContext(ctx)
}

func NewGenericConfig(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	return core2.NewGenericConfig(data, unmarshaler)
}

func ToGenericConfig(c Config) (*GenericConfig, error) {
	return core2.ToGenericConfig(c)
}

func NewConfigTypeScheme() ConfigTypeScheme {
	return core2.NewConfigTypeScheme(nil)
}

func IsGeneric(cfg Config) bool {
	return core2.IsGeneric(cfg)
}

func ErrNoContext(name string) error {
	return core2.ErrNoContext(name)
}

func IsErrNoContext(err error) bool {
	return cpi.IsErrNoContext(err)
}

func IsErrConfigNotApplicable(err error) bool {
	return cpi.IsErrConfigNotApplicable(err)
}
