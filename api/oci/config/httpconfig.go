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

The fields have the following meaning:

- <code>timeout</code> &mdash; specifies a time limit for requests made by the HTTP
  client. The timeout includes connection time, any redirects, and reading
  the response body. A timeout of zero means no timeout.

- <code>tcpDialTimeout</code> &mdash; the maximum amount of time a dial will wait
  for a TCP connect to complete. When dialing a host name with multiple IP
  addresses, the timeout may be divided between them. The operating system
  may impose its own earlier timeout.

- <code>tcpKeepAlive</code> &mdash; specifies the interval between keep-alive
  probes for an active network connection. If negative, keep-alive probes
  are disabled.

- <code>tlsHandshakeTimeout</code> &mdash; specifies the maximum amount of time
  to wait for a TLS handshake. Zero means no timeout.

- <code>responseHeaderTimeout</code> &mdash; specifies the amount of time to wait
  for a server's response headers after fully writing the request (including
  its body, if any). This time does not include the time to read the response
  body.

- <code>idleConnTimeout</code> &mdash; the maximum amount of time an idle
  (keep-alive) connection will remain idle before closing itself. Zero means
  no limit.
`
