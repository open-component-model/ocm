package internal

import (
	"encoding/json"
	"fmt"
	"strings"
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

// nonNegative requires the duration to be zero or positive.
func nonNegative(d Duration) bool {
	return !strings.HasPrefix(string(d), "-")
}

// nonNegativeOrMinusOne requires the duration to be zero, positive,
// or -1 (to disable keep-alive probes).
func nonNegativeOrMinusOne(d Duration) bool {
	return !strings.HasPrefix(string(d), "-") || strings.HasPrefix(string(d), "-1")
}

func validateDuration(name string, d *Duration, valid func(Duration) bool) error {
	if d == nil {
		return nil
	}
	if _, err := d.TimeDuration(); err != nil {
		return err
	}
	if !valid(*d) {
		return fmt.Errorf("invalid value for %s: %s", name, string(*d))
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
	// Use -1 to disable keep-alive probes.
	TCPKeepAlive *Duration `json:"tcpKeepAlive,omitempty"`

	// TLSHandshakeTimeout is the maximum time to wait for a TLS handshake.
	TLSHandshakeTimeout *Duration `json:"tlsHandshakeTimeout,omitempty"`

	// ResponseHeaderTimeout is the time limit to wait for response headers.
	ResponseHeaderTimeout *Duration `json:"responseHeaderTimeout,omitempty"`

	// IdleConnTimeout is the maximum time an idle connection remains open.
	IdleConnTimeout *Duration `json:"idleConnTimeout,omitempty"`
}

// Validate checks that timeout values are non-negative.
// TCPKeepAlive additionally allows -1 to disable keep-alive probes
// (consistent with Go's net.Dialer.KeepAlive).
func (s *HTTPSettings) Validate() error {
	for _, check := range []struct {
		name  string
		val   *Duration
		valid func(Duration) bool
	}{
		{"timeout", s.Timeout, nonNegative},
		{"tcpDialTimeout", s.TCPDialTimeout, nonNegative},
		{"tcpKeepAlive", s.TCPKeepAlive, nonNegativeOrMinusOne},
		{"tlsHandshakeTimeout", s.TLSHandshakeTimeout, nonNegative},
		{"responseHeaderTimeout", s.ResponseHeaderTimeout, nonNegative},
		{"idleConnTimeout", s.IdleConnTimeout, nonNegative},
	} {
		if err := validateDuration(check.name, check.val, check.valid); err != nil {
			return err
		}
	}
	return nil
}
