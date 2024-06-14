package rhabarber

import (
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "rhabarber.config.acme.org"
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

var (
	RhabarberType   cfgcpi.ConfigType
	RhabarberTypeV1 cfgcpi.ConfigType
)

func init() {
	RhabarberType = cfgcpi.NewConfigType[*Config](ConfigType, usage)
	cfgcpi.RegisterConfigType(RhabarberType)
	RhabarberTypeV1 = cfgcpi.NewConfigType[*Config](ConfigTypeV1, "")

	cfgcpi.RegisterConfigType(RhabarberTypeV1)
}

type Season struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Season                      `json:",inline"`
}

// NewConfig creates a new memory ConfigSpec.
func NewConfig(start, end string) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		Season: Season{
			Start: start,
			End:   end,
		},
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	t, ok := target.(*Season)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}

	*t = a.Season
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to configure the season for rhubarb:

<pre>
    type: ` + ConfigType + `
    start: mar/1
    end: apr/30
</pre>
`
