package cpi

// This is the Context Provider Interface for credential providers

import (
	"github.com/open-component-model/ocm/pkg/contexts/config/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const KIND_CONFIGTYPE = internal.KIND_CONFIGTYPE

const OCM_CONFIG_TYPE_SUFFIX = internal.OCM_CONFIG_TYPE_SUFFIX

const CONTEXT_TYPE = internal.CONTEXT_TYPE

type (
	Context          = internal.Context
	ContextProvider  = internal.ContextProvider
	Config           = internal.Config
	ConfigType       = internal.ConfigType
	ConfigTypeScheme = internal.ConfigTypeScheme
	GenericConfig    = internal.GenericConfig

	ConfigSet         = internal.ConfigSet
	ConfigurationList = internal.ConfigurationList

	ConfigApplier         = internal.ConfigApplier
	ConfigApplierFunction = internal.ConfigApplierFunction
)

var DefaultContext = internal.DefaultContext

func FromProvider(p ContextProvider) Context {
	return internal.FromProvider(p)
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

////////////////////////////////////////////////////////////////////////////////

type Updater = internal.Updater

func NewUpdater(ctx ContextProvider, target interface{}) Updater {
	return internal.NewUpdater(ctx, target)
}

func NewUpdaterForFactory[T any](ctx ContextProvider, f func() T) Updater {
	return internal.NewUpdaterForFactory(ctx, f)
}

////////////////////////////////////////////////////////////////////////////////

func ErrNoContext(name string) error {
	return internal.ErrNoContext(name)
}

func IsErrNoContext(err error) bool {
	return internal.IsErrNoContext(err)
}

func IsErrConfigNotApplicable(err error) bool {
	return internal.IsErrConfigNotApplicable(err)
}
