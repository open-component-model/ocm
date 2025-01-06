package featuregates

import (
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/datacontext/attrs/featuregatesattr"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = featuregatesattr.ATTR_SHORT + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

type FeatureGate = featuregatesattr.FeatureGate

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	featuregatesattr.Attribute  `json:",inline"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		Attribute:           *featuregatesattr.New(),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	t, ok := target.(cfgcpi.Context)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}
	if len(a.Features) == 0 {
		return nil
	}
	for n, g := range a.Features {
		featuregatesattr.SetFeature(t, n, g)
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define a list
of feature gate settings:

<pre>
    type: ` + ConfigType + `
    features:
       &lt;name>: {
          mode: off | &lt;any key to enable>
          attributes: {
             &lt;name>: &lt;any yaml value>
             ...
          }
       }
       ...
</pre>
`
