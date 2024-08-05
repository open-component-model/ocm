package empty

import (
	"ocm.software/ocm/api/oci/cpi"
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
