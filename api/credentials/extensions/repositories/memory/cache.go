package memory

import (
	"sync"

	"ocm.software/ocm/api/datacontext"
)

const ATTR_REPOS = "ocm.software/ocm/api/credentials/extensions/repositories/memory"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]*Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (r *Repositories) GetRepository(name string) *Repository {
	r.lock.Lock()
	defer r.lock.Unlock()
	repo := r.repos[name]
	if repo == nil {
		repo = NewRepository(name)
		r.repos[name] = repo
	}
	return repo
}
