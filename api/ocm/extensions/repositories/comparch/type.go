package comparch

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "ComponentArchive"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](Type, nil))
	cpi.RegisterRepositoryType(cpi.NewRepositoryType[*RepositorySpec](TypeV1, nil))
}

type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	accessio.StandardOptions    `json:",omitempty"`

	// FileFormat is the format of the repository file
	FilePath string `json:"filePath"`
	// AccessMode can be set to request readonly access or creation
	AccessMode accessobj.AccessMode `json:"accessMode,omitempty"`
}

var (
	_ accessio.Option                      = (*RepositorySpec)(nil)
	_ cpi.RepositorySpec                   = (*RepositorySpec)(nil)
	_ cpi.IntermediateRepositorySpecAspect = (*RepositorySpec)(nil)
)

// NewRepositorySpec creates a new RepositorySpec.
func NewRepositorySpec(acc accessobj.AccessMode, filePath string, opts ...accessio.Option) (*RepositorySpec, error) {
	spec := &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		FilePath:            filePath,
		AccessMode:          acc,
	}

	_, err := accessio.AccessOptions(&spec.StandardOptions, opts...)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func (a *RepositorySpec) IsIntermediate() bool {
	return true
}

func (a *RepositorySpec) GetType() string {
	return Type
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds credentials.Credentials) (cpi.Repository, error) {
	return NewRepository(ctx, a)
}

func (a *RepositorySpec) AsUniformSpec(ctx cpi.Context) *cpi.UniformRepositorySpec {
	opts := a.StandardOptions
	opts.Default(vfsattr.Get(ctx))

	p, err := vfs.Canonical(opts.GetPathFileSystem(), a.FilePath, false)
	if err != nil {
		return &cpi.UniformRepositorySpec{Type: a.GetKind(), SubPath: a.FilePath}
	}
	return &cpi.UniformRepositorySpec{Type: a.GetKind(), SubPath: p}
}

func (a *RepositorySpec) Validate(ctx cpi.Context, creds credentials.Credentials, context ...credentials.UsageContext) error {
	opts := a.StandardOptions
	opts.Default(vfsattr.Get(ctx))

	return accessobj.ValidateDescriptor(accessObjectInfo, a.FilePath, opts.GetPathFileSystem())
}
