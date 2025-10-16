package internal

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/utils/runtime"
)

const KIND_CONFIGSET = "config set"

type ConfigApplier interface {
	ApplyConfigTo(Context, cfg, tgt interface{}) error
}

type Config interface {
	runtime.VersionedTypedObject

	ApplyTo(Context, interface{}) error
}

type ConfigApplierFunction func(ctx Context, cfg, tgt interface{}) error

func (f ConfigApplierFunction) ApplyConfigTo(ctx Context, cfg, tgt interface{}) error {
	return f(ctx, cfg, tgt)
}

type ConfigSet struct {
	Description       string `json:"description,omitempty"`
	ConfigurationList `json:",inline"`
}

type ConfigurationList struct {
	Configurations []*GenericConfig `json:"configurations,omitempty"`
}

func (c *ConfigurationList) AddConfig(cfg Config) error {
	g, err := ToGenericConfig(cfg)
	if err != nil {
		return fmt.Errorf("unable to convert config to generic: %w", err)
	}

	c.Configurations = append(c.Configurations, g)

	return nil
}

func (c *ConfigurationList) AddConfigData(ctx Context, data []byte) error {
	cfg, err := ctx.GetConfigForData(data, nil)
	if err != nil {
		return errors.Wrapf(err, "invalid config specification")
	}
	g, err := ToGenericConfig(cfg)
	if err != nil {
		return fmt.Errorf("unable to convert config to generic: %w", err)
	}

	c.Configurations = append(c.Configurations, g)
	return nil
}
