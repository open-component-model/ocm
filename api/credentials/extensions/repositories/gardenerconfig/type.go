package gardenerconfig

import (
	"fmt"

	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/credentials/cpi"
	gardenercfgcpi "ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig/cpi"
	"ocm.software/ocm/api/credentials/extensions/repositories/gardenerconfig/identity"
	"ocm.software/ocm/api/credentials/internal"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "GardenerConfig"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1))
}

// RepositorySpec describes a secret server based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	URL                         string                    `json:"url"`
	ConfigType                  gardenercfgcpi.ConfigType `json:"configType"`
	Cipher                      Cipher                    `json:"cipher"`
	PropagateConsumerIdentity   *bool                     `json:"propagateConsumerIdentity,omitempty"`
}

var _ cpi.ConsumerIdentityProvider = (*RepositorySpec)(nil)

// NewRepositorySpec creates a new memory RepositorySpec.
func NewRepositorySpec(url string, configType gardenercfgcpi.ConfigType, cipher Cipher, propagateConsumerIdentity ...bool) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType:       runtime.NewVersionedTypedObject(Type),
		URL:                       url,
		ConfigType:                configType,
		Cipher:                    cipher,
		PropagateConsumerIdentity: generics.PointerTo(utils.OptionalDefaultedBool(true, propagateConsumerIdentity...)),
	}
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	r := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories)
	repos, ok := r.(*Repositories)
	if !ok {
		return nil, fmt.Errorf("failed to assert type %T to Responsitories", r)
	}

	key, err := getKey(ctx, a.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to get key from context: %w", err)
	}

	return repos.GetRepository(ctx, a.URL, a.ConfigType, a.Cipher, key, utils.AsBool(a.PropagateConsumerIdentity, true))
}

func (a *RepositorySpec) GetConsumerId(uctx ...internal.UsageContext) internal.ConsumerIdentity {
	id, err := identity.GetConsumerId(a.URL)
	if err != nil {
		return nil
	}
	return id
}

func (a *RepositorySpec) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}

func getKey(cctx cpi.Context, configURL string) ([]byte, error) {
	id, err := identity.GetConsumerId(configURL)
	if err != nil {
		return nil, err
	}

	creds, err := cpi.CredentialsForConsumer(cctx, id)
	if err != nil {
		return nil, err
	}

	var key string
	if creds != nil {
		key = creds.GetProperty(identity.ATTR_KEY)
	}

	return []byte(key), nil
}
