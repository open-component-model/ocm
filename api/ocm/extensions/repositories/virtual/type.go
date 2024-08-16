package virtual

import (
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "Virtual"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

type RepositorySpec struct {
	runtime.ObjectVersionedTypedObject
	Access Access `json:"-"`
}

func NewRepositorySpec(acc Access) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedTypedObject: runtime.NewVersionedTypedObject(Type),
		Access:                     acc,
	}
}

func (r RepositorySpec) AsUniformSpec(context cpi.Context) *cpi.UniformRepositorySpec {
	return nil
}

func (r *RepositorySpec) Repository(ctx cpi.Context, credentials credentials.Credentials) (cpi.Repository, error) {
	return NewRepository(ctx, r.Access), nil
}

var _ cpi.RepositorySpec = (*RepositorySpec)(nil)
