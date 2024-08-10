package git

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "GitRepository"
	TypeV1 = Type + runtime.VersionSeparator + "v1"

	ShortType   = "Git"
	ShortTypeV1 = ShortType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](ShortType))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](ShortTypeV1))
}

// Is checks the kind.
func Is(spec cpi.RepositorySpec) bool {
	return spec != nil && (spec.GetKind() == Type || spec.GetKind() == ShortType)
}

func IsKind(k string) bool {
	return k == Type || k == ShortType
}

// RepositorySpec describes an CTF repository interface backed by a git repository.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	accessio.StandardOptions    `json:",inline"`

	// URL is the url of the repository to resolve artifacts.
	URL string `json:"baseUrl"`

	// AccessMode can be set to request readonly access or creation
	AccessMode accessobj.AccessMode `json:"accessMode,omitempty"`

	// FileMode is the file mode for the repository in the filesystem.
	FileMode vfs.FileMode `json:"fileMode"`
}

var _ cpi.RepositorySpec = (*RepositorySpec)(nil)

var _ cpi.IntermediateRepositorySpecAspect = (*RepositorySpec)(nil)

// NewRepositorySpec creates a new RepositorySpec.
func NewRepositorySpec(mode accessobj.AccessMode, url string, fileMode vfs.FileMode, opts ...accessio.Option) (*RepositorySpec, error) {
	o, err := accessio.AccessOptions(nil, opts...)
	if err != nil {
		return nil, err
	}
	o.Default()
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		URL:                 url,
		FileMode:            fileMode,
		StandardOptions:     *o.(*accessio.StandardOptions),
		AccessMode:          mode,
	}, nil
}

func (s *RepositorySpec) IsIntermediate() bool {
	return true
}

func (s *RepositorySpec) GetType() string {
	return Type
}

func (s *RepositorySpec) Name() string {
	return s.URL
}

func (s *RepositorySpec) UniformRepositorySpec() *cpi.UniformRepositorySpec {
	u := &cpi.UniformRepositorySpec{
		Type: Type,
		Info: s.URL,
	}
	return u
}

func (s *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return New(ctx, s)
}
