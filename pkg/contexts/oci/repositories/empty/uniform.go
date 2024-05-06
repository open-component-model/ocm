package empty

import (
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
)

func init() {
	cpi.RegisterRepositorySpecHandler(&repospechandler{}, Type)
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	if u.Info != "" || u.Host == "" {
		return nil, nil
	}

	return NewRepositorySpec(), nil
}
