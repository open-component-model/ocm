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
// Negative durations are rejected because timeout values must be
// zero (disabled) or positive.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	pd, err := time.ParseDuration(str)
	if err != nil {
		return fmt.Errorf("invalid duration: %s", str)
	}
	if pd < 0 {
		return fmt.Errorf("negative duration not allowed: %s", str)
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
	TCPKeepAlive *Duration `json:"tcpKeepAlive,omitempty"`

	// TLSHandshakeTimeout is the maximum time to wait for a TLS handshake.
	TLSHandshakeTimeout *Duration `json:"tlsHandshakeTimeout,omitempty"`

	// ResponseHeaderTimeout is the time limit to wait for response headers.
	ResponseHeaderTimeout *Duration `json:"responseHeaderTimeout,omitempty"`

	// IdleConnTimeout is the maximum time an idle connection remains open.
	IdleConnTimeout *Duration `json:"idleConnTimeout,omitempty"`
}
