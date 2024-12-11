package config

import (
	"net"
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
	// entries. As a special case (for testing) it is possible
	// to configure limits for CTF, also, by using "@"+filepath.
	BlobLimits BlobLimits `json:"blobLimits"`
}

type BlobLimits map[string]int64

func (b BlobLimits) GetLimit(hostport string) int64 {
	if b == nil {
		return -1
	}
	l, ok := b[hostport]
	if ok {
		return l
	}

	if !strings.HasPrefix(hostport, "@") {
		host, _, err := net.SplitHostPort(hostport)
		if err == nil {
			l, ok = b[host]
			if ok {
				return l
			}
		}
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
blob layer limits for particular OCI registries used to host OCM repositories.
The <code>blobLimits</code> field maps a OCI registry address to the blob limit to use:

<pre>
    type: ` + ConfigType + `
    blobLimits:
        dummy.io: 65564
        dummy.io:8443: 32768 // with :8443 specifying the port and 32768 specifying the byte limit
</pre>

If blob limits apply to a registry, local blobs with a size larger than
the configured limit will be split into several layers with a maximum
size of the given value.

These settings can be overwritten by explicit settings in an OCM
repository specification for those repositories.

The most specific entry will be used. If a registry with a dedicated
port is requested, but no explicit configuration is found, the
setting for the sole hostname is used (if configured).
`
