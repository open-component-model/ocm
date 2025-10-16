package npm

import (
	"fmt"

	"github.com/mandelsoft/goutils/generics"
	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	// Type is the type of the NPMConfig.
	Type   = "NPMConfig"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1, cpi.WithDescription(usage), cpi.WithFormatSpec(format)))
}

// RepositorySpec describes a docker npmrc based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	NpmrcFile                   string `json:"npmrcFile,omitempty"`
	PropgateConsumerIdentity    *bool  `json:"propagateConsumerIdentity,omitempty"`
}

// NewRepositorySpec creates a new memory RepositorySpec.
func NewRepositorySpec(path string, propagate ...bool) *RepositorySpec {
	var p *bool
	if path == "" {
		d, err := DefaultConfig()
		if err == nil {
			path = d
		}
	}
	if len(propagate) > 0 {
		p = generics.Pointer(utils.OptionalDefaultedBool(true, propagate...))
	}

	return &RepositorySpec{
		ObjectVersionedType:      runtime.NewVersionedTypedObject(Type),
		NpmrcFile:                path,
		PropgateConsumerIdentity: p,
	}
}

func (rs *RepositorySpec) GetType() string {
	return Type
}

func (rs *RepositorySpec) Repository(ctx cpi.Context, _ cpi.Credentials) (cpi.Repository, error) {
	r := ctx.GetAttributes().GetOrCreateAttribute(".npmrc", createCache)
	cache, ok := r.(*Cache)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to Cache", r)
	}
	return cache.GetRepository(ctx, rs.NpmrcFile, utils.AsBool(rs.PropgateConsumerIdentity, true))
}
