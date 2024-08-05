package dockerconfig

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "DockerConfig"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1, cpi.WithDescription(usage), cpi.WithFormatSpec(format)))
}

// RepositorySpec describes a docker config based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	DockerConfigFile            string          `json:"dockerConfigFile,omitempty"`
	DockerConfig                json.RawMessage `json:"dockerConfig,omitempty"`
	PropgateConsumerIdentity    *bool           `json:"propagateConsumerIdentity,omitempty"`
}

func (s RepositorySpec) WithConsumerPropagation(propagate bool) *RepositorySpec {
	s.PropgateConsumerIdentity = &propagate
	return &s
}

// NewRepositorySpec creates a new memory RepositorySpec.
func NewRepositorySpec(path string, prop ...bool) *RepositorySpec {
	var p *bool
	if len(prop) > 0 {
		p = generics.Pointer(utils.Optional(prop...))
	}
	if path == "" {
		path = "~/.docker/config.json"
	}
	return &RepositorySpec{
		ObjectVersionedType:      runtime.NewVersionedTypedObject(Type),
		DockerConfigFile:         path,
		PropgateConsumerIdentity: p,
	}
}

func NewRepositorySpecForConfig(data []byte, prop ...bool) *RepositorySpec {
	var p *bool
	if len(prop) > 0 {
		p = generics.Pointer(utils.Optional(prop...))
	}
	return &RepositorySpec{
		ObjectVersionedType:      runtime.NewVersionedTypedObject(Type),
		DockerConfig:             data,
		PropgateConsumerIdentity: p,
	}
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	r := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories)
	repos, ok := r.(*Repositories)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to Repositories", r)
	}
	return repos.GetRepository(ctx, a.DockerConfigFile, a.DockerConfig, utils.AsBool(a.PropgateConsumerIdentity, true))
}
