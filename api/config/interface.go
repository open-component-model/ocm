package config

import (
	"context"

	"ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/config/internal"
	"ocm.software/ocm/api/utils/runtime"
)

const KIND_CONFIGTYPE = internal.KIND_CONFIGTYPE

const OCM_CONFIG_TYPE_SUFFIX = internal.OCM_CONFIG_TYPE_SUFFIX

const CONTEXT_TYPE = internal.CONTEXT_TYPE

var AllConfigs = internal.AllConfigs

const AllGenerations = internal.AllGenerations

type (
	Context                = internal.Context
	ContextProvider        = internal.ContextProvider
	Config                 = internal.Config
	ConfigType             = internal.ConfigType
	ConfigTypeScheme       = internal.ConfigTypeScheme
	GenericConfig          = internal.GenericConfig
	ConfigSelector         = internal.ConfigSelector
	ConfigSelectorFunction = internal.ConfigSelectorFunction
	ConfigApplier          = internal.ConfigApplier
	ConfigApplierFunction  = internal.ConfigApplierFunction
)

func DefaultContext() internal.Context {
	return internal.DefaultContext
}

func ForContext(ctx context.Context) Context {
	return internal.FromContext(ctx)
}

func FromProvider(p ContextProvider) Context {
	return internal.FromProvider(p)
}

func DefinedForContext(ctx context.Context) (Context, bool) {
	return internal.DefinedForContext(ctx)
}

func NewGenericConfig(data []byte, unmarshaler runtime.Unmarshaler) (Config, error) {
	return internal.NewGenericConfig(data, unmarshaler)
}

func ToGenericConfig(c Config) (*GenericConfig, error) {
	return internal.ToGenericConfig(c)
}

func NewConfigTypeScheme() ConfigTypeScheme {
	return internal.NewConfigTypeScheme(nil)
}

func IsGeneric(cfg Config) bool {
	return internal.IsGeneric(cfg)
}

func ErrNoContext(name string) error {
	return internal.ErrNoContext(name)
}

func IsErrNoContext(err error) bool {
	return cpi.IsErrNoContext(err)
}

func IsErrConfigNotApplicable(err error) bool {
	return cpi.IsErrConfigNotApplicable(err)
}
