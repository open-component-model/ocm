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

package helper

import (
	"io/ioutil"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Config struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Component string `json:"component"`
	Version   string `json:"version"`
}

func ReadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read config file %s", path)
	}

	var cfg Config
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse config file %s", path)
	}
	return &cfg, nil
}

func (c *Config) GetCredentials() credentials.Credentials {
	return credentials.NewCredentials(common.Properties{
		credentials.ATTR_USERNAME: c.Username,
		credentials.ATTR_PASSWORD: c.Password,
	})
}
