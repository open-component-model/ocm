package secretserver

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	SecretServerRepositoryType   = "SecretServer"
	SecretServerRepositoryTypeV1 = SecretServerRepositoryType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(SecretServerRepositoryType, cpi.NewRepositoryType(SecretServerRepositoryType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(SecretServerRepositoryTypeV1, cpi.NewRepositoryType(SecretServerRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes a secret server based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	URL                         string `json:"url"`
	ConfigName                  string `json:"configName"`
	Cipher                      Cipher `json:"cipher"`
	Key                         []byte `json:"key"`
}

// NewRepositorySpec creates a new memory RepositorySpec
func NewRepositorySpec(url string, configName string, cipher Cipher, key []byte) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(SecretServerRepositoryType),
		URL:                 url,
		ConfigName:          configName,
		Cipher:              cipher,
		Key:                 key,
	}
}

func (a *RepositorySpec) GetType() string {
	return SecretServerRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	repos := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
	return repos.GetRepository(ctx, a.URL, a.ConfigName, a.Cipher, a.Key), nil
}
