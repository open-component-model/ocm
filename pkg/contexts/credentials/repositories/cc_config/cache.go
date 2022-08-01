package cc_config

import (
	"sync"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

const ATTR_REPOS = "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/cc_config"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]*Repository
	fs    vfs.FileSystem
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (r *Repositories) GetRepository(ctx cpi.Context, url string, consumerType string, cipher Cipher, key []byte, propagate bool) *Repository {
	r.lock.Lock()
	defer r.lock.Unlock()
	repo := r.repos[url]
	if repo == nil {
		repo = NewRepository(ctx, url, consumerType, cipher, key, propagate, r.fs)
		r.repos[url] = repo
	}
	return repo
}
