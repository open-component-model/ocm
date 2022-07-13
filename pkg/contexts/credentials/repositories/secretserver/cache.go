package secretserver

import (
	"fmt"
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

const ATTR_REPOS = "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/secretserver"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]*Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (r *Repositories) GetRepository(ctx cpi.Context, url string, configName string, cipher Cipher, key []byte) *Repository {
	r.lock.Lock()
	defer r.lock.Unlock()
	id := fmt.Sprintf("%s:%s", url, configName)
	repo := r.repos[id]
	if repo == nil {
		repo = NewRepository(url, configName, cipher, key)
		r.repos[id] = repo
	}
	return repo
}
