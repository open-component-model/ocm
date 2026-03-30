package httpclient

import (
	"net"
	"net/http"
	"time"

	"ocm.software/ocm/api/oci/cpi"
)

// NewTransport creates an *http.Transport that starts as a clone of
// http.DefaultTransport and selectively overrides timeouts from cfg.
func NewTransport(cfg *cpi.HTTPSettings) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if cfg == nil {
		return transport, nil
	}

	dialTimeout, err := cfg.GetTCPDialTimeout()
	if err != nil {
		return nil, err
	}
	keepAlive, err := cfg.GetTCPKeepAlive()
	if err != nil {
		return nil, err
	}

	// TCP Dialer settings
	if dialTimeout != nil || keepAlive != nil {
		// Clone() doesn't expose the original dialer, so we create a new one
		// with the same defaults as http.DefaultTransport.
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		if dialTimeout != nil {
			dialer.Timeout = *dialTimeout
		}
		if keepAlive != nil {
			dialer.KeepAlive = *keepAlive
		}
		transport.DialContext = dialer.DialContext
	}

	tlsHandshake, err := cfg.GetTLSHandshakeTimeout()
	if err != nil {
		return nil, err
	}
	if tlsHandshake != nil {
		transport.TLSHandshakeTimeout = *tlsHandshake
	}

	responseHeader, err := cfg.GetResponseHeaderTimeout()
	if err != nil {
		return nil, err
	}
	if responseHeader != nil {
		transport.ResponseHeaderTimeout = *responseHeader
	}

	idleConn, err := cfg.GetIdleConnTimeout()
	if err != nil {
		return nil, err
	}
	if idleConn != nil {
		transport.IdleConnTimeout = *idleConn
	}

	return transport, nil
}
