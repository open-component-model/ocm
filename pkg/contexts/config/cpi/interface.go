// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

// This is the Context Provider Interface for credential providers

import (
	"github.com/open-component-model/ocm/pkg/contexts/config/core"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const KIND_CONFIGTYPE = core.KIND_CONFIGTYPE

const OCM_CONFIG_TYPE_SUFFIX = core.OCM_CONFIG_TYPE_SUFFIX

const CONTEXT_TYPE = core.CONTEXT_TYPE

type (
	Context          = core.Context
	Config           = core.Config
	ConfigType       = core.ConfigType
	ConfigTypeScheme = core.ConfigTypeScheme
	GenericConfig    = core.GenericConfig
)

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

func NewUpdater(ctx Context, target interface{}) Updater {
	return core.NewUpdater(ctx, target)
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
