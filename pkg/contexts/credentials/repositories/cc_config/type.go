package cc_config

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	CCConfigRepositoryType   = "CCConfig"
	CCConfigRepositoryTypeV1 = CCConfigRepositoryType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(CCConfigRepositoryType, cpi.NewRepositoryType(CCConfigRepositoryType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(CCConfigRepositoryTypeV1, cpi.NewRepositoryType(CCConfigRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes a secret server based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	URL                         string `json:"url"`
	ConsumerType                string `json:"consumerType"`
	Cipher                      Cipher `json:"cipher"`
	Key                         []byte `json:"key"`
	Propagate                   bool   `json:"propagate"`
}

// NewRepositorySpec creates a new memory RepositorySpec
func NewRepositorySpec(url string, consumerType string, cipher Cipher, key []byte, propagate bool) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(CCConfigRepositoryType),
		URL:                 url,
		ConsumerType:        consumerType,
		Cipher:              cipher,
		Key:                 key,
		Propagate:           propagate,
	}
}

func (a *RepositorySpec) GetType() string {
	return CCConfigRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	repos := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
	return repos.GetRepository(ctx, a.URL, a.ConsumerType, a.Cipher, a.Key, a.Propagate), nil
}
