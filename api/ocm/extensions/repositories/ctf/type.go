package ctf

import (
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const Type = ctf.Type

func NewRepositorySpec(acc accessobj.AccessMode, path string, opts ...accessio.Option) (*genericocireg.RepositorySpec, error) {
	spec, err := ctf.NewRepositorySpec(acc, path, opts...)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepositorySpec(spec, nil), nil
}
