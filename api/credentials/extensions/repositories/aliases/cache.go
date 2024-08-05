package aliases

import (
	"sync"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/datacontext"
)

const ATTR_REPOS = "ocm.software/ocm/api/credentials/extensions/repositories/aliases"

type Repositories struct {
	sync.RWMutex
	repos map[string]*Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (c *Repositories) GetRepository(name string) *Repository {
	c.RLock()
	defer c.RUnlock()
	return c.repos[name]
}

func (c *Repositories) Set(name string, spec cpi.RepositorySpec, creds cpi.CredentialsSource) {
	c.Lock()
	defer c.Unlock()
	c.repos[name] = &Repository{
		name:  name,
		spec:  spec,
		creds: creds,
	}
}
