// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"

	"github.com/open-component-model/ocm/pkg/contexts/config/core"
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const KIND_CONFIGTYPE = core.KIND_CONFIGTYPE

const OCM_CONFIG_TYPE_SUFFIX = core.OCM_CONFIG_TYPE_SUFFIX

const CONTEXT_TYPE = core.CONTEXT_TYPE

var AllConfigs = core.AllConfigs

const AllGenerations = core.AllGenerations

type (
	Context                = core.Context
	Config                 = core.Config
	ConfigType             = core.ConfigType
	ConfigTypeScheme       = core.ConfigTypeScheme
	GenericConfig          = core.GenericConfig
	ConfigSelector         = core.ConfigSelector
	ConfigSelectorFunction = core.ConfigSelectorFunction
)

func DefaultContext() core.Context {
	return core.DefaultContext
}

func ForContext(ctx context.Context) Context {
	return core.ForContext(ctx)
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	return core.DefinedForContext(ctx)
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

func ErrNoContext(name string) error {
	return core.ErrNoContext(name)
}

func IsErrNoContext(err error) bool {
	return cpi.IsErrNoContext(err)
}

func IsErrConfigNotApplicable(err error) bool {
	return cpi.IsErrConfigNotApplicable(err)
}
