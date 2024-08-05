package docker

import (
	"ocm.software/ocm/api/oci/cpi"
)

func init() {
	cpi.RegisterRepositorySpecHandler(&repospechandler{}, Type)
}

type repospechandler struct{}

func (h *repospechandler) MapReference(ctx cpi.Context, u *cpi.UniformRepositorySpec) (cpi.RepositorySpec, error) {
	host := u.Host
	if u.Scheme != "" && host != "" {
		host = u.Scheme + "://" + u.Host
	}
	if u.Info != "" {
		if u.Info == "default" {
			host = ""
		} else if host == "" {
			host = u.Info
		}
	}
	return NewRepositorySpec(host), nil
}
