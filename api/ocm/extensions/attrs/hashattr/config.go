package hashattr

import (
	"github.com/mandelsoft/goutils/errors"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
	"ocm.software/ocm/api/utils/listformat"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "hasher" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1, usage))
}

// Config describes a memory based repository interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	HashAlgorithm               string `json:"hashAlgorithm"`
}

// New creates a new memory ConfigSpec.
func New(algo string) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		HashAlgorithm:       algo,
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	t, ok := target.(Context)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}
	return errors.Wrapf(t.GetAttributes().SetAttribute(ATTR_KEY, a.HashAlgorithm), "applying config failed")
}

var usage = `
The config type <code>` + ConfigType + `</code> can be used to define
the default hash algorithm used to calculate digests for resources.
It supports the field <code>hashAlgorithm</code>, with one of the following
values:
` + listformat.FormatList(sha256.Algorithm, signing.DefaultRegistry().HasherNames()...)
