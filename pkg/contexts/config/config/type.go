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
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "generic" + cpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterConfigType(ConfigType, cpi.NewConfigType(ConfigType, &Config{}, usage))
	cpi.RegisterConfigType(ConfigTypeV1, cpi.NewConfigType(ConfigTypeV1, &Config{}, usage))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Configurations              []*cpi.GenericConfig `json:"configurations"`
}

// NewConfigSpec creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(ConfigType),
		Configurations:      []*cpi.GenericConfig{},
	}
}

func (c *Config) AddConfig(cfg cpi.Config) error {
	g, err := cpi.ToGenericConfig(cfg)
	if err != nil {
		return fmt.Errorf("unable to convert cpi config to generic: %w", err)
	}

	c.Configurations = append(c.Configurations, g)

	return nil
}

func (c *Config) GetType() string {
	return ConfigType
}

func (c *Config) ApplyTo(ctx cpi.Context, target interface{}) error {
	if cctx, ok := target.(cpi.Context); ok {
		list := errors.ErrListf("applying generic config list")
		for i, cfg := range c.Configurations {
			sub := fmt.Sprintf("config entry %d", i)
			list.Add(cctx.ApplyConfig(cfg, sub))
		}
		return list.Result()
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define a list
of arbitrary configuration specifications:

<pre>
    type: ` + ConfigType + `
    configurations:
      - type: &lt;any config type>
        ...
      ...
</pre>
`
