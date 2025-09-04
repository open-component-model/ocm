package config

import (
	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/preferrelativeattr"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "oci.uploader" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1))
}

// Config describes a memory based config interface.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	UploadOptions
}

// New creates a new memory ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
	}
}

func (a *Config) ApplyTo(ctx config.Context, target interface{}) error {
	t, ok := target.(*UploadOptions)
	if !ok {
		return config.ErrNoContext(ConfigType)
	}
	t.Repositories = append(t.Repositories, a.Repositories...)
	t.PreferRelativeAccess = a.PreferRelativeAccess
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to set some
configurations for the implicit OCI artifact upload for OCI based OCM repositories.

<pre>
    type: ` + ConfigType + `
    preferRelativeAccess: true # use relative access methods for given target repositories.
    repositories:
	- localhost:5000
</pre>

If <code>preferRelativeAccess</code> is set to <code>true</code> the 
OCI uploader for OCI based OCM repositories does not use the
OCI repository to create absolute OCI access methods
if the target repository is in the <code>repositories</code> list.
Instead, a relative  <code>relativeOciReference</code> access method 
is created.
If this list is empty, all uploads are handled this way.

If the global attribute <code>` + preferrelativeattr.ATTR_SHORT + `</code> 
is configured, it overrides the <code>preferRelativeAccess</code> setting.
`
