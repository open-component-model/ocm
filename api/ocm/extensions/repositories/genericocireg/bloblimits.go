package genericocireg

import (
	"sync"

	configctx "ocm.software/ocm/api/config"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/config"
)

var (
	defaultBlobLimits config.BlobLimits
	lock              sync.Mutex
)

const (
	KB = int64(1000)
	MB = 1000 * KB
	GB = 1000 * MB
)

func init() {
	defaultBlobLimits = config.BlobLimits{}

	// Add limits for known OCI repositories, here,
	// or provide init functions in specialized packages
	// by calling AddDefaultBlobLimit.
	AddDefaultBlobLimit("ghcr.io", 10*GB) // https://github.com/orgs/community/discussions/77429
}

// AddDefaultBlobLimit can be used to set default blob limits
// for known repositories.
// Those limits will be overwritten, by blob limits
// given by a configuration object and the repository
// specification.
func AddDefaultBlobLimit(name string, limit int64) {
	lock.Lock()
	defer lock.Unlock()

	defaultBlobLimits[name] = limit
}

func ConfigureBlobLimits(ctx configctx.ContextProvider, target config.Configurable) {
	if target != nil {
		lock.Lock()
		defer lock.Unlock()

		target.ConfigureBlobLimits(defaultBlobLimits)

		if ctx != nil {
			ctx.ConfigContext().ApplyTo(0, target)
		}
	}
}
