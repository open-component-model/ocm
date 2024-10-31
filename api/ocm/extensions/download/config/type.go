package config

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "downloader.ocm" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based config interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	Registrations               []Registration `json:"registrations,omitempty"`
}

type Registration struct {
	Name                    string `json:"name"`
	Description             string `json:"description,omitempty"`
	download.HandlerOptions `json:",inline"`
	Config                  download.HandlerConfig `json:"config,omitempty"`
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(ConfigType),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) AddRegistration(hdlrs ...Registration) error {
	for i, h := range hdlrs {
		if h.Name == "" {
			return fmt.Errorf("handler registration %d requires a name", i)
		}
	}
	a.Registrations = append(a.Registrations, hdlrs...)
	return nil
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	t, ok := target.(cpi.Context)
	if !ok {
		return config.ErrNoContext(ConfigType)
	}
	reg := download.For(t)
	for _, h := range a.Registrations {
		accepted, err := reg.RegisterByName(h.Name, t, h.Config, &h.HandlerOptions)
		if err != nil {
			return errors.Wrapf(err, "registering download handler %q[%s]", h.Name, h.Description)
		}
		if !accepted {
			download.Logger(t).Info("no matching handler for configuration %q[%s]", h.Name, h.Description)
		}
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define a list
of preconfigured download handler registrations (see <CMD>ocm ocm-downloadhandlers</CMD>):

<pre>
    type: ` + ConfigType + `
    description: "my standard download handler configuration"
    registrations:
      - name: oci/artifact
        artifactType: ociImage
        mimeType: ...
        description: ...
        priority: ...
        config: ...
      ...
</pre>
`
