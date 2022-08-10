package gardenerconfig

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	gardenercfg_cpi "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

const ATTR_REPOS = "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]*Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (r *Repositories) GetRepository(ctx cpi.Context, url string, configType gardenercfg_cpi.ConfigType, cipher Cipher, key []byte, propagateConsumerIdentity bool) (*Repository, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	repo := r.repos[url]
	if repo == nil {
		repo, err := NewRepository(ctx, url, configType, cipher, key, propagateConsumerIdentity)
		if err != nil {
			return nil, err
		}
		r.repos[url] = repo
	}
	return repo, nil
}
