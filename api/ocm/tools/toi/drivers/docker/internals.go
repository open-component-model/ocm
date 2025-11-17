package docker

import (
	"net"
	"strings"

	"github.com/distribution/reference"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/moby/moby/api/types/registry"
	registrytypes "github.com/moby/moby/api/types/registry"
)

const (
	// DefaultNamespace is the default namespace
	DefaultNamespace = "docker.io"
	// IndexHostname is the index hostname, used for authentication and image search.
	IndexHostname = "index.docker.io"
	// IndexServer is used for user auth and image search
	IndexServer = "https://index.docker.io/v1/"
	// IndexName is the name of the index
	IndexName = "docker.io"
)

// NewIndexInfo creates a new [registry.IndexInfo] or the given
// repository-name, and detects whether the registry is considered
// "secure" (non-localhost).
func NewIndexInfo(reposName reference.Named) *registry.IndexInfo {
	indexName := normalizeIndexName(reference.Domain(reposName))
	if indexName == IndexName {
		return &registry.IndexInfo{
			Name:     IndexName,
			Secure:   true,
			Official: true,
		}
	}

	return &registry.IndexInfo{
		Name:   indexName,
		Secure: !isInsecure(indexName),
	}
}

func normalizeIndexName(val string) string {
	if val == "index.docker.io" {
		return "docker.io"
	}
	return val
}

// isInsecure is used to detect whether a registry domain or IP-address is allowed
// to use an insecure (non-TLS, or self-signed cert) connection according to the
// defaults, which allows for insecure connections with registries running on a
// loopback address ("localhost", "::1/128", "127.0.0.0/8").
//
// It is used in situations where we don't have access to the daemon's configuration,
// for example, when used from the client / CLI.
func isInsecure(hostNameOrIP string) bool {
	// Attempt to strip port if present; this also strips brackets for
	// IPv6 addresses with a port (e.g. "[::1]:5000").
	//
	// This is best-effort; we'll continue using the address as-is if it fails.
	if host, _, err := net.SplitHostPort(hostNameOrIP); err == nil {
		hostNameOrIP = host
	}
	if hostNameOrIP == "127.0.0.1" || hostNameOrIP == "::1" || strings.EqualFold(hostNameOrIP, "localhost") {
		// Fast path; no need to resolve these, assuming nobody overrides
		// "localhost" for anything else than a loopback address (sorry, not sorry).
		return true
	}

	var addresses []net.IP
	if ip := net.ParseIP(hostNameOrIP); ip != nil {
		addresses = append(addresses, ip)
	} else {
		// Try to resolve the host's IP-addresses.
		addrs, _ := net.LookupIP(hostNameOrIP)
		addresses = append(addresses, addrs...)
	}

	for _, addr := range addresses {
		if addr.IsLoopback() {
			return true
		}
	}
	return false
}

// authConfigKey is the key used to store credentials for Docker Hub. It is
// a copy of [registry.IndexServer].
//
// [registry.IndexServer]: https://pkg.go.dev/github.com/docker/docker@v28.3.3+incompatible/registry#IndexServer
const authConfigKey = "https://index.docker.io/v1/"

// ResolveAuthConfig returns auth-config for the given registry from the
// credential-store. It returns an empty AuthConfig if no credentials were
// found.
func resolveAuthConfig(cfg *configfile.ConfigFile, index *registrytypes.IndexInfo) registrytypes.AuthConfig {
	configKey := index.Name
	if index.Official {
		configKey = authConfigKey
	}

	a, _ := cfg.GetAuthConfig(configKey)
	return registrytypes.AuthConfig{
		Username:      a.Username,
		Password:      a.Password,
		ServerAddress: a.ServerAddress,

		// TODO(thaJeztah): Are these expected to be included?
		Auth:          a.Auth,
		IdentityToken: a.IdentityToken,
		RegistryToken: a.RegistryToken,
	}
}
