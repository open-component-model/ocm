package plugin

import (
	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/config/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
)

var _ cpi.Config = (*Config)(nil)

type Config struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

func (c *Config) ApplyTo(context internal.Context, i interface{}) error {
	return nil
}

func New(name string, desc string) cpi.ConfigType {
	return cpi.NewConfigType[*Config](name, desc)
}
