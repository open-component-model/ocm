package httptimeoutattr

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mandelsoft/goutils/errors"

	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType         = "http" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1Alpha1 = ConfigType + runtime.VersionSeparator + "v1alpha1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, configUsage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1Alpha1, configUsage))
}

// Timeout wraps time.Duration to support JSON/YAML marshaling
// of both human-readable duration strings (e.g. "30s", "5m", "1h")
// and nanosecond numbers.
type Timeout time.Duration

func (d Timeout) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Timeout) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("failed to parse HTTP client timeout: %w", err)
	}

	switch value := v.(type) {
	case float64:
		*d = Timeout(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid timeout value %q: must be a duration like 30s, 5m, or nanoseconds number: %w", value, err)
		}
		*d = Timeout(tmp)
		return nil
	default:
		return fmt.Errorf("timeout must be a duration string or nanoseconds number, got %T", v)
	}
}

// Config describes the configuration for HTTP client settings.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Timeout                     Timeout `json:"timeout,omitempty"`
}

// NewConfig creates a new HTTP config with the given timeout.
func NewConfig(timeout time.Duration) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		Timeout:             Timeout(timeout),
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
	if a.Timeout != 0 {
		return errors.Wrapf(t.GetAttributes().SetAttribute(ATTR_KEY, time.Duration(a.Timeout)), "applying config failed")
	}
	return nil
}

const configUsage = `
The config type <code>` + ConfigType + `</code> can be used to configure
HTTP client settings:

<pre>
    type: ` + ConfigType + `
    timeout: 30s
</pre>

The <code>timeout</code> field specifies the HTTP client timeout as a
Go duration string (e.g. "30s", "5m", "1h"). If not set, the default
timeout of 30s is used.
`
