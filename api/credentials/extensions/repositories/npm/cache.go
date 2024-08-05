package npm

import (
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/datacontext"
)

type Cache struct {
	repos map[string]*Repository
}

func createCache(_ datacontext.Context) interface{} {
	return &Cache{
		repos: map[string]*Repository{},
	}
}

func (r *Cache) GetRepository(ctx cpi.Context, name string, prop bool) (*Repository, error) {
	var (
		err  error = nil
		repo *Repository
	)
	if name != "" {
		repo = r.repos[name]
	}
	if repo == nil {
		repo, err = NewRepository(ctx, name, prop)
		if err == nil {
			r.repos[name] = repo
		}
	}
	return repo, err
}
