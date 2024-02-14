package npm

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

type Cache struct {
	repos map[string]*Repository
}

func createCache(_ datacontext.Context) interface{} {
	return &Cache{
		repos: map[string]*Repository{},
	}
}

func (r *Cache) GetRepository(ctx cpi.Context, name string) (*Repository, error) {
	var (
		err  error = nil
		repo *Repository
	)
	if name != "" {
		repo = r.repos[name]
	}
	if repo == nil {
		repo, err = NewRepository(ctx, name)
		if err == nil {
			r.repos[name] = repo
		}
	}
	return repo, err
}
