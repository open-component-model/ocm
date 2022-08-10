package gardenerconfig

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	gardenercfg_cpi "github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/gardenerconfig/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	RepositoryType   = "GardenerConfig"
	RepositoryTypeV1 = RepositoryType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(RepositoryType, cpi.NewRepositoryType(RepositoryType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(RepositoryTypeV1, cpi.NewRepositoryType(RepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes a secret server based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	URL                         string                     `json:"url"`
	ConfigType                  gardenercfg_cpi.ConfigType `json:"configType"`
	Cipher                      Cipher                     `json:"cipher"`
	Key                         []byte                     `json:"key"`
	PropagateConsumerIdentity   bool                       `json:"propagateConsumerIdentity"`
}

// NewRepositorySpec creates a new memory RepositorySpec
func NewRepositorySpec(url string, configType gardenercfg_cpi.ConfigType, cipher Cipher, key []byte, propagateConsumerIdentity bool) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType:       runtime.NewVersionedObjectType(RepositoryType),
		URL:                       url,
		ConfigType:                configType,
		Cipher:                    cipher,
		Key:                       key,
		PropagateConsumerIdentity: propagateConsumerIdentity,
	}
}

func (a *RepositorySpec) GetType() string {
	return RepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	repos := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
	return repos.GetRepository(ctx, a.URL, a.ConfigType, a.Cipher, a.Key, a.PropagateConsumerIdentity)
}
