package attrs

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"

	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "attributes" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	// Attributes describe a set of generic attribute settings
	Attributes map[string]json.RawMessage `json:"attributes,omitempty"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		Attributes:          map[string]json.RawMessage{},
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) AddAttribute(attr string, value interface{}) error {
	data, err := datacontext.DefaultAttributeScheme.Encode(attr, value, runtime.DefaultJSONEncoding)
	if err == nil {
		a.Attributes[attr] = data
	}
	return err
}

func (a *Config) AddRawAttribute(attr string, data []byte) error {
	_, err := datacontext.DefaultAttributeScheme.Decode(attr, data, runtime.DefaultJSONEncoding)
	if err == nil {
		a.Attributes[attr] = data
	}
	return err
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	list := errors.ErrListf("applying config")
	t, ok := target.(cfgcpi.Context)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}
	if a.Attributes == nil {
		return nil
	}
	for a, e := range a.Attributes {
		eff := datacontext.DefaultAttributeScheme.Shortcuts()[a]
		if eff != "" {
			a = eff
		}
		list.Add(errors.Wrapf(t.GetAttributes().SetEncodedAttribute(a, e, runtime.DefaultJSONEncoding), "attribute %q", a))
	}
	return list.Result()
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define a list
of arbitrary attribute specifications:

<pre>
    type: ` + ConfigType + `
    attributes:
       &lt;name>: &lt;yaml defining the attribute>
       ...
</pre>
`
