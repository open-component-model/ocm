package artifactset

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "ArtifactSet"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1))
}

const (
	FORMAT_OCI = "oci/v1"
	FORMAT_OCM = "ocm/v1"
)

type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	Options                     `json:",inline"`

	// FileFormat is the format of the repository file
	FilePath string `json:"filePath"`
	// AccessMode can be set to request readonly access or creation
	AccessMode accessobj.AccessMode `json:"accessMode,omitempty"`

	FormatVersion string `json:"formatVersion,omitempty"`
}

// NewRepositorySpec creates a new RepositorySpec.
func NewRepositorySpec(acc accessobj.AccessMode, filePath string, opts ...accessio.Option) (*RepositorySpec, error) {
	o, err := accessio.AccessOptions(&Options{}, opts...)
	if err != nil {
		return nil, err
	}
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		FilePath:            filePath,
		Options:             *o.(*Options),
		AccessMode:          acc,
	}, nil
}

func (s *RepositorySpec) Name() string {
	return s.FilePath
}

func (s *RepositorySpec) GetFormatVersion() string {
	if s.FormatVersion == "" {
		return FORMAT_OCM
	}
	return s.FormatVersion
}

func (s *RepositorySpec) UniformRepositorySpec() *cpi.UniformRepositorySpec {
	u := &cpi.UniformRepositorySpec{
		Type: Type,
		Info: s.FilePath,
	}
	return u
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return NewRepository(ctx, a)
}

func (a *RepositorySpec) AsUniformSpec(ctx cpi.Context) cpi.UniformRepositorySpec {
	opts, _ := NewOptions(&a.Options) // now unknown option possible (same Options type)
	opts.Default(vfsattr.Get(ctx))
	p, err := vfs.Canonical(opts.GetPathFileSystem(), a.FilePath, false)
	if err != nil {
		return cpi.UniformRepositorySpec{Type: a.GetKind(), Info: a.FilePath}
	}
	return cpi.UniformRepositorySpec{Type: a.GetKind(), Info: p}
}

func (a *RepositorySpec) Validate(ctx cpi.Context, creds credentials.Credentials, context ...credentials.UsageContext) error {
	opts := a.Options
	opts.Default(vfsattr.Get(ctx))

	return accessobj.ValidateDescriptor(&baseInfo, a.FilePath, opts.GetPathFileSystem())
}
