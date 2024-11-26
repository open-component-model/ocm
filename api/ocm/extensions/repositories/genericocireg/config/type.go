package config

import (
	"strings"

	"ocm.software/ocm/api/config"
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	ConfigType   = "blobLimits.ocireg.ocm" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1))
}

// Config describes a memory based config interface
// for configuring blob limits for underlying OCI manifest layers.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	// BlobLimits describe the limit setting for host:port
	// entries. As a spcial case (for testing) it is possible
	// to configure linits for CTF, also, by using "@"+filepath.
	BlobLimits BlobLimits `json:"blobLimits"`
}

type BlobLimits map[string]int64

func (b BlobLimits) GetLimit(hostport string) int64 {
	if b == nil {
		return -1
	}
	host := hostport
	i := strings.Index(hostport, ":")
	if i > 0 {
		host = hostport[:i]
	}

	l, ok := b[hostport]
	if ok {
		return l
	}
	l, ok = b[host]
	if ok {
		return l
	}
	return -1
}

type Configurable interface {
	ConfigureBlobLimits(limits BlobLimits)
}

// New creates a blob limit ConfigSpec.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) AddLimit(hostport string, limit int64) {
	if a.BlobLimits == nil {
		a.BlobLimits = BlobLimits{}
	}
	a.BlobLimits[hostport] = limit
}

func (a *Config) ApplyTo(ctx config.Context, target interface{}) error {
	t, ok := target.(Configurable)
	if !ok {
		return config.ErrNoContext(ConfigType)
	}
	if a.BlobLimits != nil {
		t.ConfigureBlobLimits(a.BlobLimits)
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to set some
configurations for an OCM context;

<pre>
    type: ` + ConfigType + `
    aliases:
       myrepo: 
          type: &lt;any repository type>
          &lt;specification attributes>
          ...
    resolvers:
      - repository:
          type: &lt;any repository type>
          &lt;specification attributes>
          ...
        prefix: ghcr.io/open-component-model/ocm
        priority: 10
</pre>

With aliases repository alias names can be mapped to a repository specification.
The alias name can be used in a string notation for an OCM repository.

Resolvers define a list of OCM repository specifications to be used to resolve
dedicated component versions. These settings are used to compose a standard
component version resolver provided for an OCM context. Optionally, a component
name prefix can be given. It limits the usage of the repository to resolve only
components with the given name prefix (always complete name segments).
An optional priority can be used to influence the lookup order. Larger value
means higher priority (default 10).

All matching entries are tried to lookup a component version in the following
order:
- highest priority first
- longest matching sequence of component name segments first.

If resolvers are defined, it is possible to use component version names on the
command line without a repository. The names are resolved with the specified
resolution rule.
They are also used as default lookup repositories to lookup component references
for recursive operations on component versions (<code>--lookup</code> option).
`
