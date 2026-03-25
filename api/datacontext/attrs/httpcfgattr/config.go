package httpcfgattr

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
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigType, usage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*Config](ConfigTypeV1Alpha1, usage))
}

// Duration is a string type representing a Go duration (e.g. "30s", "5m").
// It is validated on JSON unmarshaling.
type Duration string

// UnmarshalJSON implements the json.Unmarshaller interface.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	if _, err := time.ParseDuration(str); err != nil {
		return fmt.Errorf("invalid duration: %s", str)
	}
	*d = Duration(str)
	return nil
}

// TimeDuration parses the Duration string and returns a time.Duration.
// Returns 0 if the string is empty or invalid.
func (d *Duration) TimeDuration() time.Duration {
	pd, _ := time.ParseDuration(string(*d))
	return pd
}

// NewDuration creates a pointer to a Duration.
func NewDuration(d time.Duration) *Duration {
	v := Duration(d.String())
	return &v
}

// HTTPSettings contains the timeout settings for HTTP clients.
// All timeout values use Duration (Go duration strings in config).
// If not set (nil), the http.DefaultTransport value from the Go
// standard library is used.
//
// Note: Timeout controls the overall http.Client deadline and is
// independent of the transport-level timeouts below. Setting Timeout
// alone does NOT override TCPDialTimeout, TLSHandshakeTimeout, etc.
type HTTPSettings struct {
	// Timeout is the overall http.Client timeout — the maximum duration
	// for the entire request including connection, TLS, headers, and body.
	// It does NOT serve as a fallback for transport-level timeouts.
	// If not set, http.Client uses no timeout (0).
	Timeout *Duration `json:"timeout,omitempty"`

	// TCPDialTimeout is the time limit for establishing a TCP connection.
	TCPDialTimeout *Duration `json:"tcpDialTimeout,omitempty"`

	// TCPKeepAlive is the interval between TCP keep-alive probes.
	TCPKeepAlive *Duration `json:"tcpKeepAlive,omitempty"`

	// TLSHandshakeTimeout is the maximum time to wait for a TLS handshake.
	TLSHandshakeTimeout *Duration `json:"tlsHandshakeTimeout,omitempty"`

	// ResponseHeaderTimeout is the time limit to wait for response headers.
	ResponseHeaderTimeout *Duration `json:"responseHeaderTimeout,omitempty"`

	// IdleConnTimeout is the maximum time an idle connection remains open.
	IdleConnTimeout *Duration `json:"idleConnTimeout,omitempty"`
}

// GetTimeout returns the overall HTTP client timeout.
// Returns 0 (disabled) if not set.
func (s *HTTPSettings) GetTimeout() time.Duration {
	if s == nil || s.Timeout == nil {
		return 0
	}
	return s.Timeout.TimeDuration()
}

// Config describes the configuration for HTTP client settings.
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	HTTPSettings                `json:",inline"`
}

// New creates a new empty HTTP Config.
func New() *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

// NewConfig creates a new HTTP config with the given overall timeout.
func NewConfig(timeout time.Duration) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedTypedObject(ConfigType),
		HTTPSettings: HTTPSettings{
			Timeout: NewDuration(timeout),
		},
	}
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	if t, ok := target.(Context); ok {
		if t.AttributesContext().IsAttributesContext() { // apply only to root context
			return errors.Wrapf(a.ApplyToAttribute(Get(t)), "applying config failed")
		}
	}
	return cfgcpi.ErrNoContext(ConfigType)
}

// ApplyToAttribute merges this config's settings into an existing attribute.
func (a *Config) ApplyToAttribute(attr *Attribute) error {
	s := &attr.settings
	if a.Timeout != nil {
		s.Timeout = a.Timeout
	}
	if a.TCPDialTimeout != nil {
		s.TCPDialTimeout = a.TCPDialTimeout
	}
	if a.TCPKeepAlive != nil {
		s.TCPKeepAlive = a.TCPKeepAlive
	}
	if a.TLSHandshakeTimeout != nil {
		s.TLSHandshakeTimeout = a.TLSHandshakeTimeout
	}
	if a.ResponseHeaderTimeout != nil {
		s.ResponseHeaderTimeout = a.ResponseHeaderTimeout
	}
	if a.IdleConnTimeout != nil {
		s.IdleConnTimeout = a.IdleConnTimeout
	}
	return nil
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to configure
HTTP client settings:

<pre>
    type: ` + ConfigType + `
    timeout: 0s
    tcpDialTimeout: 30s
    tcpKeepAlive: 30s
    tlsHandshakeTimeout: 10s
    responseHeaderTimeout: 0s
    idleConnTimeout: 90s
</pre>

All timeout values are Go duration strings (e.g. "30s", "5m", "1h").
Use "0s" to disable a specific timeout. If not set, the <code>http.DefaultTransport</code>
values from the Go standard library are used.

Note: <code>timeout</code> controls the overall <code>http.Client</code> request deadline and is
independent of the transport-level settings. Setting only <code>timeout</code> does not
affect <code>tcpDialTimeout</code>, <code>tlsHandshakeTimeout</code>, or other transport timeouts.
`
