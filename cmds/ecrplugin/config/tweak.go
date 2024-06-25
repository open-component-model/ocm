package config

import (
	"github.com/open-component-model/ocm/api/ocm/plugin/ppi"
)

func TweakDescriptor(d ppi.Descriptor, cfg *Config) ppi.Descriptor {
	if cfg != nil {
		d.Actions[0].DefaultSelectors = append(d.Actions[0].DefaultSelectors, cfg.Hostnames...)
	}
	return d
}
