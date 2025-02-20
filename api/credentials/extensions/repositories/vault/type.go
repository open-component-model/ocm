package vault

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/credentials/extensions/repositories/vault/identity"
	"ocm.software/ocm/api/credentials/internal"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "HashiCorpVault"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1, cpi.WithDescription(usage), cpi.WithFormatSpec(format)))
}

// RepositorySpec describes a docker config based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	ServerURL                   string `json:"serverURL"`
	Options                     `json:",inline"`
}

var _ cpi.ConsumerIdentityProvider = (*RepositorySpec)(nil)

// NewRepositorySpec creates a new memory RepositorySpec.
func NewRepositorySpec(url string, opts ...Option) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		ServerURL:           url,
		Options:             *optionutils.EvalOptions(opts...),
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
	spec := *a
	spec.Secrets = slices.Clone(a.Secrets)
	if spec.MountPath == "" {
		spec.MountPath = "secret"
	}
	return repos.GetRepository(ctx, &spec)
}

func (a *RepositorySpec) GetKey() cpi.ProviderIdentity {
	spec := *a
	spec.PropgateConsumerIdentity = false
	data, err := json.Marshal(&spec)
	if err == nil {
		return cpi.ProviderIdentity(data)
	}
	return cpi.ProviderIdentity(spec.ServerURL)
}

func (a *RepositorySpec) GetConsumerId(uctx ...internal.UsageContext) internal.ConsumerIdentity {
	id, err := identity.GetConsumerId(a.ServerURL, a.Namespace, a.MountPath, a.Path)
	if err != nil {
		return nil
	}
	return id
}

func (a *RepositorySpec) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}
