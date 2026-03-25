package httpclient

import (
	"net"
	"net/http"
	"time"

	"ocm.software/ocm/api/datacontext/attrs/httpcfgattr"
)

// NewTransport creates an *http.Transport that starts as a clone of
// http.DefaultTransport and selectively overrides timeouts from cfg.
func NewTransport(cfg *httpcfgattr.HTTPSettings) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if cfg == nil {
		return transport
	}

	// TCP Dialer settings
	if cfg.TCPDialTimeout != nil || cfg.TCPKeepAlive != nil {
		// Clone() doesn't expose the original dialer, so we create a new one
		// with the same defaults as http.DefaultTransport.
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		if cfg.TCPDialTimeout != nil {
			dialer.Timeout = cfg.TCPDialTimeout.TimeDuration()
		}
		if cfg.TCPKeepAlive != nil {
			dialer.KeepAlive = cfg.TCPKeepAlive.TimeDuration()
		}
		transport.DialContext = dialer.DialContext
	}

	if cfg.TLSHandshakeTimeout != nil {
		transport.TLSHandshakeTimeout = cfg.TLSHandshakeTimeout.TimeDuration()
	}
	if cfg.ResponseHeaderTimeout != nil {
		transport.ResponseHeaderTimeout = cfg.ResponseHeaderTimeout.TimeDuration()
	}
	if cfg.IdleConnTimeout != nil {
		transport.IdleConnTimeout = cfg.IdleConnTimeout.TimeDuration()
	}
	return transport
}
