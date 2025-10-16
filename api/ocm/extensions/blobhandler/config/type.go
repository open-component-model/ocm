package config

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/blobhandler"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "uploader.ocm" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
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
	Name                       string `json:"name"`
	Description                string `json:"description,omitempty"`
	blobhandler.HandlerOptions `json:",inline"`
	Config                     blobhandler.HandlerConfig `json:"config,omitempty"`
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
	reg := blobhandler.For(t)
	for _, h := range a.Registrations {
		opts := h.HandlerOptions
		if opts.Priority == 0 {
			// config objects have higher prio than builtin defaults
			// CLI options get even higher prio.
			opts.Priority = blobhandler.DEFAULT_BLOBHANDLER_PRIO * 2
		}
		accepted, err := reg.RegisterByName(h.Name, t, h.Config, &opts)
		if err != nil {
			return errors.Wrapf(err, "registering upload handler %q[%s]", h.Name, h.Description)
		}
		if !accepted {
			download.Logger(t).Info("no matching handler for configuration %q[%s]", h.Name, h.Description)
		}
	}
	return nil
}

var usage = `
The config type <code>` + ConfigType + `</code> can be used to define a list
of preconfigured upload handler registrations (see <CMD>ocm ocm-uploadhandlers</CMD>),
the default priority is ` + fmt.Sprintf("%d", download.DEFAULT_BLOBHANDLER_PRIO*2) + `:

<pre>
    type: ` + ConfigType + `
    description: "my standard upload handler configuration"
    registrations:
      - name: oci/artifact
        artifactType: ociImage
        config:
          ociRef: ghcr.io/open-component-model/...
      ...
</pre>
`
