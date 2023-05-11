// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "plugin" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(ConfigType, cfgcpi.NewConfigType(ConfigType, &Config{}, usage))
	cfgcpi.RegisterConfigType(ConfigTypeV1, cfgcpi.NewConfigType(ConfigTypeV1, &Config{}, usage))
}

// Config describes a memory based config interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Plugin                      string          `json:"plugin"`
	Config                      json.RawMessage `json:"config"`
}

// New creates a new memory ConfigSpec.
func New(name string, data []byte) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		Plugin:              name,
		Config:              data,
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) ApplyTo(ctx config.Context, target interface{}) error {
	t, ok := target.(Target)
	if !ok {
		return config.ErrNoContext(ConfigType)
	}
	t.ConfigurePlugin(a.Plugin, a.Config)
	return nil
}

type Target interface {
	ConfigurePlugin(name string, config json.RawMessage)
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to configure a 
plugin.

<pre>
    type: ` + ConfigType + `
    plugin: &lt;plugin name>
    config: &lt;arbitrary configuration structure>
</pre>
`
