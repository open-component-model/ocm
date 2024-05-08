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
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Component  string `json:"component,omitempty"`
	Repository string `json:"repository,omitempty"`
	Version    string `json:"version,omitempty"`

	Target    json.RawMessage `json:"targetRepository,omitempty"`
	OCMConfig string          `json:"ocmConfig,omitempty"`
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
