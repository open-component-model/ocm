package config

import (
	"encoding/json"

	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "plugin" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based config interface for plugins.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Plugin                      string          `json:"plugin"`
	Config                      json.RawMessage `json:"config,omitempty"`
	DisableAutoRegistration     bool            `json:"disableAutoRegistration,omitempty"`
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
	t.DisableAutoConfiguration(a.Plugin, a.DisableAutoRegistration)
	return nil
}

type Target interface {
	ConfigurePlugin(name string, config json.RawMessage)
	DisableAutoConfiguration(name string, flag bool)
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to configure a 
plugin.

<pre>
    type: ` + ConfigType + `
    plugin: &lt;plugin name>
    config: &lt;arbitrary configuration structure>
    disableAutoRegistration: &lt;boolean flag to disable auto registration for up- and download handlers>
</pre>
`
