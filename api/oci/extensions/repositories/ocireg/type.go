package ocireg

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/containerd/containerd/reference"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/tech/oci/identity"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/utils/tcp"
)

const (
	LegacyType = "ociRegistry"
	Type       = "OCIRegistry"
	TypeV1     = Type + runtime.VersionSeparator + "v1"

	ShortType   = "oci"
	ShortTypeV1 = ShortType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](LegacyType))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](ShortType))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](ShortTypeV1))
}

// Is checks the kind.
func Is(spec cpi.RepositorySpec) bool {
	return spec != nil && spec.GetKind() == Type || spec.GetKind() == LegacyType
}

func IsKind(k string) bool {
	return k == Type || k == LegacyType
}

// RepositorySpec describes an OCI registry interface backed by an oci registry.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	// BaseURL is the base url of the repository to resolve artifacts.
	BaseURL     string `json:"baseUrl"`
	LegacyTypes *bool  `json:"legacyTypes,omitempty"`
}

var (
	_ cpi.RepositorySpec                   = (*RepositorySpec)(nil)
	_ credentials.ConsumerIdentityProvider = (*RepositorySpec)(nil)
)

// NewRepositorySpec creates a new RepositorySpec.
func NewRepositorySpec(baseURL string) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		BaseURL:             baseURL,
	}
}

func NewLegacyRepositorySpec(baseURL string) *RepositorySpec {
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(LegacyType),
		BaseURL:             baseURL,
	}
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Name() string {
	return a.BaseURL
}

func (a *RepositorySpec) UniformRepositorySpec() *cpi.UniformRepositorySpec {
	return cpi.UniformRepositorySpecForHostURL(Type, a.BaseURL)
}

func (a *RepositorySpec) getInfo(creds credentials.Credentials) (*RepositoryInfo, error) {
	var u *url.URL
	info := &RepositoryInfo{}
	legacy := false
	ref, err := reference.Parse(a.BaseURL)
	if err == nil {
		u, err = url.Parse("https://" + ref.Locator)
		if err != nil {
			return nil, err
		}
		info.Locator = ref.Locator
		if ref.Object != "" {
			return nil, fmt.Errorf("invalid repository locator %q", a.BaseURL)
		}
	} else {
		u, err = url.Parse(a.BaseURL)
		if err != nil {
			return nil, err
		}
		info.Locator = u.Host
	}
	if a.LegacyTypes != nil {
		legacy = *a.LegacyTypes
	} else {
		idx := strings.Index(info.Locator, "/")
		host := info.Locator
		if idx > 0 {
			host = info.Locator[:idx]
		}
		if host == "docker.io" {
			legacy = true
		}
	}
	info.Scheme = u.Scheme
	info.Creds = creds
	info.Legacy = legacy

	return info, nil
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	info, err := a.getInfo(creds)
	if err != nil {
		return nil, err
	}
	return NewRepository(ctx, a, info)
}

func (a *RepositorySpec) Validate(ctx cpi.Context, creds credentials.Credentials, context ...credentials.UsageContext) error {
	info, err := a.getInfo(creds)
	if err != nil {
		return err
	}
	h, p, _ := info.HostInfo()
	if p == "" {
		p = "443"
	}
	return tcp.PingTCPServer(h+":"+p, time.Second)
}

func (a *RepositorySpec) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	info, err := a.getInfo(nil)
	if err != nil {
		return nil
	}
	if c, ok := utils.Optional(uctx...).(credentials.StringUsageContext); ok {
		return identity.GetConsumerId(info.Locator, c.String())
	}
	return identity.GetConsumerId(info.Locator, "")
}

func (a *RepositorySpec) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}
