package config

import (
	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "oci" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based config interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Aliases                     map[string]*cpi.GenericRepositorySpec `json:"aliases,omitempty"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) SetAlias(name string, spec cpi.RepositorySpec) error {
	g, err := cpi.ToGenericRepositorySpec(spec)
	if err != nil {
		return err
	}
	if a.Aliases == nil {
		a.Aliases = map[string]*cpi.GenericRepositorySpec{}
	}
	a.Aliases[name] = g
	return nil
}

func (a *Config) ApplyTo(ctx config.Context, target interface{}) error {
	t, ok := target.(cpi.Context)
	if !ok {
		return config.ErrNoContext(ConfigType)
	}
	for n, s := range a.Aliases {
		t.SetAlias(n, s)
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define
OCI registry aliases:

<pre>
    type: ` + ConfigType + `
    aliases:
       &lt;name>: &lt;OCI registry specification>
       ...
</pre>
`
