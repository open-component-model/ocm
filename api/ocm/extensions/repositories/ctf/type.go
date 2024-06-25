package ctf

import (
	"github.com/open-component-model/ocm/api/oci/extensions/repositories/ctf"
	"github.com/open-component-model/ocm/api/ocm/extensions/repositories/genericocireg"
	"github.com/open-component-model/ocm/api/utils/accessio"
	"github.com/open-component-model/ocm/api/utils/accessobj"
)

const Type = ctf.Type

func NewRepositorySpec(acc accessobj.AccessMode, path string, opts ...accessio.Option) (*genericocireg.RepositorySpec, error) {
	spec, err := ctf.NewRepositorySpec(acc, path, opts...)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepositorySpec(spec, nil), nil
}
