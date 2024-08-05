package dockerconfig

import (
	"sync"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/datacontext"
)

const ATTR_REPOS = "ocm.software/ocm/api/credentials/extensions/repositories/dockerconfig"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]*Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (r *Repositories) GetRepository(ctx cpi.Context, name string, data []byte, propagate bool) (*Repository, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	var (
		err  error = nil
		repo *Repository
	)
	if name != "" {
		repo = r.repos[name]
	}
	if repo == nil {
		repo, err = NewRepository(ctx, name, data, propagate)
		if err == nil {
			r.repos[name] = repo
		}
	}
	return repo, err
}
