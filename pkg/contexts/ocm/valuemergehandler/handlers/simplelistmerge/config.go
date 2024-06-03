package simplelistmerge

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
)

func NewConfig(fields ...string) *Config {
	return &Config{IgnoredFields: fields}
}

type Config struct {
	IgnoredFields []string `json:"ignoredFields,omitempty"`
}

var _ hpi.Config = (*Config)(nil)

func (c *Config) Complete(ctx hpi.Context) error {
	return nil
}
