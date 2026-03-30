package config

import (
	cfgcpi "ocm.software/ocm/api/config/cpi"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	HTTPConfigType         = "http" + cfgcpi.OCM_CONFIG_TYPE_SUFFIX
	HTTPConfigTypeV1Alpha1 = HTTPConfigType + runtime.VersionSeparator + "v1alpha1"
)

func init() {
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*HTTPConfig](HTTPConfigType, httpUsage))
	cfgcpi.RegisterConfigType(cfgcpi.NewConfigType[*HTTPConfig](HTTPConfigTypeV1Alpha1, httpUsage))
}

// HTTPConfig describes the configuration for HTTP client settings.
type HTTPConfig struct {
	runtime.ObjectVersionedType `json:",inline"`
	cpi.HTTPSettings            `json:",inline"`
}

func (a *HTTPConfig) GetType() string {
	return HTTPConfigType
}

func (a *HTTPConfig) ApplyTo(_ cfgcpi.Context, target interface{}) error {
	t, ok := target.(cpi.Context)
	if !ok {
		return cfgcpi.ErrNoContext(HTTPConfigType)
	}
	s, err := t.GetHTTPSettings()
	if err != nil {
		return err
	}
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
	t.SetHTTPSettings(&s)
	return nil
}

const httpUsage = `
The config type <code>` + HTTPConfigType + `</code> can be used to configure
HTTP client settings:

<pre>
    type: generic.config.ocm.software/v1
    configurations:
      - type: ` + HTTPConfigType + `
        timeout: "0s"
        tcpDialTimeout: "30s"
        tcpKeepAlive: "30s"
        tlsHandshakeTimeout: "10s"
        responseHeaderTimeout: "0s"
        idleConnTimeout: "90s"
</pre>

All timeout values are Go duration strings (e.g. "30s", "5m", "1h").
Use "0s" to disable a specific timeout. If not set, the <code>http.DefaultTransport</code>
values from the Go standard library are used.

Note: <code>timeout</code> controls the overall <code>http.Client</code> request deadline and is
independent of the transport-level settings. Setting only <code>timeout</code> does not
affect <code>tcpDialTimeout</code>, <code>tlsHandshakeTimeout</code>, or other transport timeouts.
`
