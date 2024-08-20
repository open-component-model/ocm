package git

import (
	"ocm.software/ocm/api/oci/extensions/repositories/git"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const Type = git.Type

func NewRepositorySpec(acc accessobj.AccessMode, url string, opts ...accessio.Option) (*genericocireg.RepositorySpec, error) {
	spec, err := git.NewRepositorySpec(acc, url, opts...)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepositorySpec(spec, nil), nil
}
