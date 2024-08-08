package composition

import (
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "Composition"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type, nil))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1, nil))
}

type RepositorySpec struct {
	runtime.ObjectVersionedTypedObject
	Name string `json:"name"`
}

var _ cpi.RepositorySpec = (*RepositorySpec)(nil)

func NewRepositorySpec(name string) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedTypedObject: runtime.NewVersionedTypedObject(Type),
		Name:                       name,
	}
}

func (r RepositorySpec) AsUniformSpec(context cpi.Context) *cpi.UniformRepositorySpec {
	return nil
}

func (r *RepositorySpec) Repository(ctx cpi.Context, credentials credentials.Credentials) (cpi.Repository, error) {
	return NewRepository(ctx, r.Name), nil
}

var _ cpi.RepositorySpec = (*RepositorySpec)(nil)
