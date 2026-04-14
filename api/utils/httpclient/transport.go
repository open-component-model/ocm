package httpclient

import (
	"net"
	"net/http"
	"time"

	"ocm.software/ocm/api/oci/cpi"
)

// Default dialer timeouts matching http.DefaultTransport.
const (
	defaultDialTimeout = 30 * time.Second
	defaultKeepAlive   = 30 * time.Second
)

// NewTransport creates an *http.Transport that starts as a clone of
// http.DefaultTransport and selectively overrides timeouts from cfg.
func NewTransport(cfg *cpi.HTTPSettings) *http.Transport {
	dt, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		dt = &http.Transport{}
	}
	transport := dt.Clone()

	if cfg == nil {
		return transport
	}

	// TCP Dialer settings
	if cfg.TCPDialTimeout != nil || cfg.TCPKeepAlive != nil {
		// Clone() doesn't expose the original dialer, so we create a new one
		// with the same defaults as http.DefaultTransport.
		dialer := &net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: defaultKeepAlive,
		}
		if cfg.TCPDialTimeout != nil {
			dialer.Timeout = time.Duration(*cfg.TCPDialTimeout)
		}
		if cfg.TCPKeepAlive != nil {
			dialer.KeepAlive = time.Duration(*cfg.TCPKeepAlive)
		}
		transport.DialContext = dialer.DialContext
	}

	if cfg.TLSHandshakeTimeout != nil {
		transport.TLSHandshakeTimeout = time.Duration(*cfg.TLSHandshakeTimeout)
	}

	if cfg.ResponseHeaderTimeout != nil {
		transport.ResponseHeaderTimeout = time.Duration(*cfg.ResponseHeaderTimeout)
	}

	if cfg.IdleConnTimeout != nil {
		transport.IdleConnTimeout = time.Duration(*cfg.IdleConnTimeout)
	}

	return transport
}
