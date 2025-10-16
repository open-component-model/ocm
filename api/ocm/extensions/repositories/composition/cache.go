package composition

import (
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/refmgmt"
)

const ATTR_REPOS = "ocm.software/ocm/api/ocm/extensions/repositories/composition"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]cpi.Repository
}

var _ finalizer.Finalizable = (*Repositories)(nil)

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]cpi.Repository{},
	}
}

func (r *Repositories) GetRepository(name string) cpi.Repository {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.repos[name]
}

func (r *Repositories) SetRepository(name string, repo cpi.Repository) {
	r.lock.Lock()
	defer r.lock.Unlock()

	old := r.repos[name]
	if old != nil {
		refmgmt.AsLazy(old).Close()
	}
	r.repos[name] = repo
}

func (r *Repositories) Finalize() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	list := errors.ErrListf("composition repositories")
	for n, r := range r.repos {
		list.Addf(nil, refmgmt.AsLazy(r).Close(), "repository %s", n)
	}

	r.repos = map[string]cpi.Repository{}
	return list.Result()
}

func Cleanup(ctx cpi.ContextProvider) error {
	repos := ctx.OCMContext().GetAttributes().GetAttribute(ATTR_REPOS)
	if repos != nil {
		return repos.(*Repositories).Finalize()
	}
	return nil
}
