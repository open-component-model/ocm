package simplelistmerge

import (
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
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
