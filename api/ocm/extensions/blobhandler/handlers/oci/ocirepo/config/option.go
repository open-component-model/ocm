package config

import (
	"net"
)

// UploadOptions is used to configure
// the implicit OCI uploader for a local OCM repository.
// It can be used to request the generation of relative
// OCI access methods, generally or for dedicated targets.
type UploadOptions struct {
	// PreferRelativeAccess enables or disables the settings.
	PreferRelativeAccess bool `json:"preferRelativeAccess,omitempty"`
	// Repositories is list of repository specs, with or without port.
	// If no filters are configured all repos are matched.
	Repositories []string `json:"repositories,omitempty"`
}

// PreferRelativeAccessFor checks a repo spec for using
// a relative access method instead of an absolute one.
// It checks hostname and optionally a port name.
// The most specific configuration wins.
func (o *UploadOptions) PreferRelativeAccessFor(repo string) bool {
	if len(o.Repositories) == 0 || !o.PreferRelativeAccess {
		return o.PreferRelativeAccess
	}

	fallback := false

	host, port, err := net.SplitHostPort(repo)
	if err != nil {
		host = repo
	}
	for _, r := range o.Repositories {
		rhost, rport, err := net.SplitHostPort(r)
		if err != nil {
			rhost = r
		}
		if host == rhost {
			if rport == "" {
				fallback = true
			}
			if port == rport {
				return true
			}
		}
	}
	return fallback
}
