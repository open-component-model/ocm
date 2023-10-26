// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"encoding/json"
	"io/ioutil"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Config struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Component  string `json:"component"`
	Repository string `json:"repository"`
	Version    string `json:"version"`

	Target    json.RawMessage `json:"targetRepository"`
	OCMConfig json.RawMessage `json:"ocmConfig"`
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
	return identity.SimpleCredentials(c.Username, c.Password)
}
