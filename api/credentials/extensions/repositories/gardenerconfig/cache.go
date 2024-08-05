package gardenerconfig

import (
	"fmt"
	"sync"

	"ocm.software/ocm/api/credentials/cpi"
	gardenercfgcpi "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig/cpi"
	"ocm.software/ocm/api/datacontext"
)

const ATTR_REPOS = "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]*Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (r *Repositories) GetRepository(ctx cpi.Context, url string, configType gardenercfgcpi.ConfigType, cipher Cipher, key []byte, propagateConsumerIdentity bool) (*Repository, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.repos[url]; !ok {
		repo, err := NewRepository(ctx, url, configType, cipher, key, propagateConsumerIdentity)
		if err != nil {
			return nil, fmt.Errorf("unable to create repository: %w", err)
		}
		r.repos[url] = repo
	}
	return r.repos[url], nil
}
