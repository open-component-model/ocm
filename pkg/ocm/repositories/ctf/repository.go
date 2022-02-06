package ctf

import (
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf/comparch"
)

type Repository struct {
	ctx  cpi.Context
	spec *RepositorySpec
	arch *comparch.ComponentArchive
}

var _ cpi.Repository = (*Repository)(nil)

func NewRepository(ctx cpi.Context, s *RepositorySpec) (*Repository, error) {
	r := &Repository{ctx, s, nil}
	a, err := r.Open()
	if err != nil {
		return nil, err
	}
	r.arch = a
	return r, err
}

func (r *Repository) Get() *comparch.ComponentArchive {
	if r.arch != nil {
		return r.arch
	}
	return nil
}

func (r *Repository) Open() (*comparch.ComponentArchive, error) {
	a, err := comparch.Open(r.ctx, r.spec.AccessMode, r.spec.FilePath, 0700, r.spec.Options)
	if err != nil {
		return nil, err
	}
	r.arch = a
	return a, nil
}

func (r *Repository) GetContext() core.Context {
	return r.ctx
}

func (r *Repository) GetSpecification() core.RepositorySpec {
	return r.spec
}

func (r *Repository) ExistsComponentVersion(name string, ref string) (bool, error) {
	if r.arch == nil {
		return false, accessio.ErrClosed
	}
	return r.arch.GetName() == name && r.arch.GetVersion() == ref, nil
}

func (r *Repository) LookupComponentVersion(name string, version string) (cpi.ComponentVersionAccess, error) {
	ok, err := r.ExistsComponentVersion(name, version)
	if ok {
		return r.arch, nil
	}
	return nil, err
}

func (r *Repository) LookupComponent(name string) (cpi.ComponentAccess, error) {
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if r.arch.GetName() != name {
		return nil, errors.ErrNotFound(errors.KIND_COMPONENT, name, CTFComponentArchiveType)
	}
	return &ComponentAccess{r}, nil
}

func (r Repository) Close() error {
	if r.arch != nil {
		r.arch.Close()
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type ComponentAccess struct {
	repo *Repository
}

var _ cpi.ComponentAccess = (*ComponentAccess)(nil)

func (c ComponentAccess) GetContext() cpi.Context {
	return c.repo.GetContext()
}

func (c ComponentAccess) GetVersion(ref string) (cpi.ComponentVersionAccess, error) {
	return c.repo.LookupComponentVersion(c.repo.arch.GetName(), ref)
}

func (c ComponentAccess) AddVersion(access core.ComponentVersionAccess) error {
	return errors.ErrNotSupported(errors.KIND_FUNCTION, "add version", CTFComponentArchiveType)
}

func (c ComponentAccess) NewVersion(version string) (core.ComponentVersionAccess, error) {
	return nil, errors.ErrNotSupported(errors.KIND_FUNCTION, "new version", CTFComponentArchiveType)
}
