// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/common"
	cpi2 "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	GenericConfigType   = "generic.config" + common.TypeGroupSuffix
	GenericConfigTypeV1 = GenericConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi2.RegisterConfigType(GenericConfigType, cpi2.NewConfigType(GenericConfigType, &Config{}))
	cpi2.RegisterConfigType(GenericConfigTypeV1, cpi2.NewConfigType(GenericConfigTypeV1, &Config{}))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Configurations              []*cpi2.GenericConfig `json:"configurations"`
}

// NewConfig creates a new memory Config
func NewConfig(info string) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(GenericConfigType),
	}
}

func (c *Config) AddConfig(cfg cpi2.Config) error {
	g, err := cpi2.ToGenericConfig(cfg)
	if err != nil {
		return err
	}
	c.Configurations = append(c.Configurations, g)
	return nil
}

func (c *Config) GetType() string {
	return GenericConfigType
}

func (c *Config) ApplyTo(ctx cpi2.Context, target interface{}) error {
	if cctx, ok := target.(cpi2.Context); ok {
		list := errors.ErrListf("applying generic config list")
		for i, cfg := range c.Configurations {
			sub := fmt.Sprintf("config entry %d", i)
			list.Add(cctx.ApplyConfig(cfg, sub))
		}
		return list.Result()
	}
	return nil
}
