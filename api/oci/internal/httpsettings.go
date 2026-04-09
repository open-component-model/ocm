package internal

import (
	"encoding/json"
	"fmt"
	"time"
)

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

// TimeDuration parses the Duration string and returns a *time.Duration.
// Returns (nil, nil) if d is nil or empty — callers must distinguish
// nil (not configured) from zero (explicitly disabled).
// Returns an error if the string is malformed.
func (d *Duration) TimeDuration() (*time.Duration, error) {
	if d == nil || *d == "" {
		return nil, nil
	}
	pd, err := time.ParseDuration(string(*d))
	if err != nil {
		return nil, fmt.Errorf("invalid duration %q: %w", string(*d), err)
	}
	return &pd, nil
}

func validateNonNegative(name string, d *Duration) error {
	if d == nil {
		return nil
	}
	td, err := d.TimeDuration()
	if err != nil {
		return err
	}
	if td != nil && *td < 0 {
		return fmt.Errorf("invalid value for %s: %s, must be zero or positive", name, string(*d))
	}
	return nil
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
	// Timeout is the overall http.Client timeout -- the maximum duration
	// for the entire request including connection, TLS, headers, and body.
	// It does NOT serve as a fallback for transport-level timeouts.
	// If not set, http.Client uses no timeout (0).
	Timeout *Duration `json:"timeout,omitempty"`

	// TCPDialTimeout is the time limit for establishing a TCP connection.
	TCPDialTimeout *Duration `json:"tcpDialTimeout,omitempty"`

	// TCPKeepAlive is the interval between TCP keep-alive probes.
	// If zero, probes are sent with a default value (currently 15 seconds).
	// If negative, keep-alive probes are disabled.
	TCPKeepAlive *Duration `json:"tcpKeepAlive,omitempty"`

	// TLSHandshakeTimeout is the maximum time to wait for a TLS handshake.
	TLSHandshakeTimeout *Duration `json:"tlsHandshakeTimeout,omitempty"`

	// ResponseHeaderTimeout is the time limit to wait for response headers.
	ResponseHeaderTimeout *Duration `json:"responseHeaderTimeout,omitempty"`

	// IdleConnTimeout is the maximum time an idle connection remains open.
	IdleConnTimeout *Duration `json:"idleConnTimeout,omitempty"`
}

// Validate checks that timeout values are non-negative.
// TCPKeepAlive is not validated because any negative value
// disables keep-alive probes (consistent with Go's net.Dialer.KeepAlive).
func (s *HTTPSettings) Validate() error {
	for _, check := range []struct {
		name string
		val  *Duration
	}{
		{"timeout", s.Timeout},
		{"tcpDialTimeout", s.TCPDialTimeout},
		{"tlsHandshakeTimeout", s.TLSHandshakeTimeout},
		{"responseHeaderTimeout", s.ResponseHeaderTimeout},
		{"idleConnTimeout", s.IdleConnTimeout},
	} {
		if err := validateNonNegative(check.name, check.val); err != nil {
			return err
		}
	}
	return nil
}
