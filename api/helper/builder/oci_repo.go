package builder

import (
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
)

const T_OCIREPOSITORY = "oci repository"

type ociRepository struct {
	base
	kind string
	cpi.Repository
}

func (r *ociRepository) Type() string {
	if r.kind != "" {
		return r.kind
	}
	return T_OCIREPOSITORY
}

func (r *ociRepository) Set() {
	r.Builder.oci_repo = r.Repository
}

func (b *Builder) GeneralOCIRepository(spec oci.RepositorySpec, f ...func()) {
	repo, err := b.OCIContext().RepositoryForSpec(spec)
	b.failOn(err)
	b.configure(&ociRepository{Repository: repo, kind: T_OCIREPOSITORY}, f)
}

func (b *Builder) OCIRegistry(url string, path string, f ...func()) {
	spec := ocireg.NewRepositorySpec(url)
	b.GeneralOCIRepository(spec, f...)
}
