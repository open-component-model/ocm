package genericocireg

import (
	"encoding/json"
	"path"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/sirupsen/logrus"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/componentmapping"
	"ocm.software/ocm/api/utils/runtime"
)

// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
// to OCI Image References.
type ComponentNameMapping string

const (
	Type = ocireg.Type

	OCIRegistryURLPathMapping ComponentNameMapping = "urlPath"
	OCIRegistryDigestMapping  ComponentNameMapping = "sha256-digest"
)

func init() {
	cpi.DefaultDelegationRegistry().Register("OCI", New(10))
}

// delegation tries to resolve an ocm repository specification
// with an OCI repository specification supported by the OCI context
// of the OCM context.
type delegation struct {
	prio int
}

func New(prio int) cpi.RepositoryPriorityDecoder {
	return &delegation{prio}
}

var _ cpi.RepositoryPriorityDecoder = (*delegation)(nil)

func (d *delegation) Decode(ctx cpi.Context, data []byte, unmarshal runtime.Unmarshaler) (cpi.RepositorySpec, error) {
	if unmarshal == nil {
		unmarshal = runtime.DefaultYAMLEncoding.Unmarshaler
	}

	ospec, err := ctx.OCIContext().RepositoryTypes().Decode(data, unmarshal)
	if err != nil {
		return nil, err
	}
	if oci.IsUnknown(ospec) {
		return nil, nil
	}

	meta := &ComponentRepositoryMeta{}
	err = unmarshal.Unmarshal(data, meta)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal component repository meta information")
	}
	return normalizers.Normalize(NewRepositorySpec(ospec, meta)), nil
}

func (d *delegation) Priority() int {
	return d.prio
}

// ComponentRepositoryMeta describes config special for a mapping of
// a component repository to an oci registry.
// It is parsed in addition to an OCI based specification.
type ComponentRepositoryMeta struct {
	// ComponentNameMapping describes the method that is used to map the "Component Name", "Component Version"-tuples
	// to OCI Image References.
	ComponentNameMapping ComponentNameMapping `json:"componentNameMapping,omitempty"`
	SubPath              string               `json:"subPath,omitempty"`
}

func NewComponentRepositoryMeta(subPath string, mapping ComponentNameMapping) *ComponentRepositoryMeta {
	return DefaultComponentRepositoryMeta(&ComponentRepositoryMeta{
		ComponentNameMapping: mapping,
		SubPath:              subPath,
	})
}

////////////////////////////////////////////////////////////////////////////////

type RepositorySpec struct {
	oci.RepositorySpec
	ComponentRepositoryMeta
	BlobLimit *int64
}

var (
	_ cpi.RepositorySpec                   = (*RepositorySpec)(nil)
	_ cpi.PrefixProvider                   = (*RepositorySpec)(nil)
	_ cpi.IntermediateRepositorySpecAspect = (*RepositorySpec)(nil)
	_ json.Marshaler                       = (*RepositorySpec)(nil)
	_ credentials.ConsumerIdentityProvider = (*RepositorySpec)(nil)
)

func NewRepositorySpec(spec oci.RepositorySpec, meta *ComponentRepositoryMeta) *RepositorySpec {
	s := &RepositorySpec{
		RepositorySpec:          spec,
		ComponentRepositoryMeta: *DefaultComponentRepositoryMeta(meta),
	}
	return normalizers.Normalize(s)
}

func (a *RepositorySpec) PathPrefix() string {
	return a.SubPath
}

func (a *RepositorySpec) IsIntermediate() bool {
	if s, ok := a.RepositorySpec.(oci.IntermediateRepositorySpecAspect); ok {
		return s.IsIntermediate()
	}
	return false
}

// TODO: host etc is missing

func (a *RepositorySpec) AsUniformSpec(cpi.Context) *cpi.UniformRepositorySpec {
	spec := a.RepositorySpec.UniformRepositorySpec()
	return &cpi.UniformRepositorySpec{Type: a.GetKind(), Scheme: spec.Scheme, Host: spec.Host, Info: spec.Info, TypeHint: spec.TypeHint, SubPath: a.SubPath}
}

type meta struct {
	ComponentRepositoryMeta `json:",inline"`
	BlobLimit               *int64 `json:"blobLimit,omitempty"`
}

func (u *RepositorySpec) UnmarshalJSON(data []byte) error {
	logrus.Debugf("unmarshal generic ocireg spec %s\n", string(data))
	ocispec := &oci.GenericRepositorySpec{}
	if err := json.Unmarshal(data, ocispec); err != nil {
		return err
	}

	var m meta
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	u.RepositorySpec = ocispec
	u.ComponentRepositoryMeta = m.ComponentRepositoryMeta
	if m.BlobLimit != nil {
		u.BlobLimit = m.BlobLimit
	}

	normalizers.Normalize(u)
	return nil
}

// MarshalJSON implements a custom json unmarshal method for an unstructured type.
// The oci.RepositorySpec object might already implement json.Marshaler,
// which would be inherited and omit marshaling the addend attributes of a
// cpi.RepositorySpec.
func (u RepositorySpec) MarshalJSON() ([]byte, error) {
	ocispec, err := runtime.ToUnstructuredTypedObject(u.RepositorySpec)
	if err != nil {
		return nil, err
	}

	m := meta{
		ComponentRepositoryMeta: u.ComponentRepositoryMeta,
		BlobLimit:               u.BlobLimit,
	}
	compmeta, err := runtime.ToUnstructuredObject(&m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(compmeta.FlatMerge(ocispec.Object))
}

func (s *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	r, err := s.RepositorySpec.Repository(ctx.OCIContext(), creds)
	if err != nil {
		return nil, err
	}
	if s.BlobLimit != nil {
		return NewRepository(ctx, &s.ComponentRepositoryMeta, r, *s.BlobLimit), nil
	}
	return NewRepository(ctx, &s.ComponentRepositoryMeta, r), nil
}

func (s *RepositorySpec) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	prefix := s.SubPath
	if c, ok := general.Optional(uctx...).(credentials.StringUsageContext); ok {
		prefix = path.Join(prefix, componentmapping.ComponentDescriptorNamespace, c.String())
	}
	return credentials.GetProvidedConsumerId(s.RepositorySpec, credentials.StringUsageContext(prefix))
}

func (s *RepositorySpec) GetIdentityMatcher() string {
	return credentials.GetProvidedIdentityMatcher(s.RepositorySpec)
}

func (s *RepositorySpec) Validate(ctx cpi.Context, creds credentials.Credentials, uctx ...credentials.UsageContext) error {
	return s.RepositorySpec.Validate(ctx.OCIContext(), creds, uctx...)
}

func DefaultComponentRepositoryMeta(meta *ComponentRepositoryMeta) *ComponentRepositoryMeta {
	if meta == nil {
		meta = &ComponentRepositoryMeta{}
	}
	if meta.ComponentNameMapping == "" {
		meta.ComponentNameMapping = OCIRegistryURLPathMapping
	}
	return meta
}
