// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comparch

import (
	"fmt"
	"strings"
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessio/refmgmt"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	ocmhdlr "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/support"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

type _RepositoryImplBase = cpi.RepositoryImplBase

type RepositoryImpl struct {
	_RepositoryImplBase
	lock sync.RWMutex
	arch *ComponentArchive
}

var _ cpi.RepositoryImpl = (*RepositoryImpl)(nil)

func NewRepository(ctx cpi.Context, s *RepositorySpec) (cpi.Repository, error) {
	if s.GetPathFileSystem() == nil {
		s.SetPathFileSystem(vfsattr.Get(ctx))
	}
	a, err := Open(ctx, s.AccessMode, s.FilePath, 0o700, s)
	if err != nil {
		return nil, err
	}
	return a.AsRepository(), nil
}

func newRepository(a *ComponentArchive) (main cpi.Repository, nonref cpi.Repository) {
	// close main cv abstraction on repository close -------v
	base := cpi.NewRepositoryImplBase(a.GetContext(), a.ComponentVersionAccess)
	impl := &RepositoryImpl{
		_RepositoryImplBase: *base,
		arch:                a,
	}
	return cpi.NewRepository(impl), cpi.NewNoneRefRepositoryView(impl)
}

func (r *RepositoryImpl) ComponentLister() cpi.ComponentLister {
	return r
}

func (r *RepositoryImpl) matchPrefix(prefix string, closure bool) bool {
	if r.arch.GetName() != prefix {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if !closure || !strings.HasPrefix(r.arch.GetName(), prefix) {
			return false
		}
	}
	return true
}

func (r *RepositoryImpl) NumComponents(prefix string) (int, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return -1, accessio.ErrClosed
	}
	if !r.matchPrefix(prefix, true) {
		return 0, nil
	}
	return 1, nil
}

func (r *RepositoryImpl) GetComponents(prefix string, closure bool) ([]string, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if !r.matchPrefix(prefix, closure) {
		return []string{}, nil
	}
	return []string{r.arch.GetName()}, nil
}

func (r *RepositoryImpl) Get() *ComponentArchive {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch != nil {
		return r.arch
	}
	return nil
}

func (r *RepositoryImpl) GetSpecification() cpi.RepositorySpec {
	return r.arch.spec
}

func (r *RepositoryImpl) ExistsComponentVersion(name string, ref string) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return false, accessio.ErrClosed
	}
	return r.arch.GetName() == name && r.arch.GetVersion() == ref, nil
}

func (r *RepositoryImpl) LookupComponentVersion(name string, version string) (cpi.ComponentVersionAccess, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	ok, err := r.ExistsComponentVersion(name, version)
	if !ok {
		if err == nil {
			err = errors.ErrNotFound(cpi.KIND_COMPONENTVERSION, common.NewNameVersion(name, version).String(), Type)
		}
		return nil, err
	}
	c, err := newComponentAccess(r)
	if err != nil {
		return nil, err
	}
	defer refmgmt.PropagateCloseTemporary(&err, c) // temporary component object not exposed.
	return c.LookupVersion(version)
}

func (r *RepositoryImpl) LookupComponent(name string) (cpi.ComponentAccess, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if r.arch.GetName() != name {
		return nil, errors.ErrNotFound(errors.KIND_COMPONENT, name, Type)
	}
	return newComponentAccess(r)
}

////////////////////////////////////////////////////////////////////////////////

type _ComponentAccessImplBase = cpi.ComponentAccessImplBase

type ComponentAccessImpl struct {
	_ComponentAccessImplBase
	repo *RepositoryImpl
}

var _ cpi.ComponentAccessImpl = (*ComponentAccessImpl)(nil)

func newComponentAccess(r *RepositoryImpl) (cpi.ComponentAccess, error) {
	base, err := cpi.NewComponentAccessImplBase(r.GetContext(), r.arch.GetName(), r)
	if err != nil {
		return nil, err
	}
	impl := &ComponentAccessImpl{
		_ComponentAccessImplBase: *base,
		repo:                     r,
	}
	return cpi.NewComponentAccess(impl, "component archive"), nil
}

func (c *ComponentAccessImpl) IsReadOnly() bool {
	return c.repo.arch.IsReadOnly()
}

func (c *ComponentAccessImpl) ListVersions() ([]string, error) {
	return []string{c.repo.arch.GetVersion()}, nil
}

func (c *ComponentAccessImpl) HasVersion(vers string) (bool, error) {
	return vers == c.repo.arch.GetVersion(), nil
}

func (c *ComponentAccessImpl) LookupVersion(version string) (cpi.ComponentVersionAccess, error) {
	if version != c.repo.arch.GetVersion() {
		return nil, errors.ErrNotFound(cpi.KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", c.GetName(), c.repo.arch.GetVersion()))
	}
	return newComponentVersionAccess(c, version, false)
}

func (c *ComponentAccessImpl) container(access cpi.ComponentVersionAccess) *componentArchiveContainer {
	mine, _ := support.GetComponentVersionContainer[*ComponentVersionContainer](access)
	if mine == nil || mine.comp != c {
		return nil
	}
	return mine.comp.repo.arch.container
}

func (c *ComponentAccessImpl) IsOwned(access cpi.ComponentVersionAccess) bool {
	return c.container(access) == c.repo.arch.container
}

func (c *ComponentAccessImpl) AddVersion(access cpi.ComponentVersionAccess) error {
	if access.GetName() != c.GetName() {
		return errors.ErrInvalid("component name", access.GetName())
	}
	mine := c.container(access)
	if mine == nil {
		return errors.Newf("component version not owned by component archive")
	}
	return nil
}

func (c *ComponentAccessImpl) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	if version != c.repo.arch.GetVersion() {
		return nil, errors.ErrNotSupported(cpi.KIND_COMPONENTVERSION, version, fmt.Sprintf("component archive %s:%s", c.GetName(), c.repo.arch.GetVersion()))
	}
	if !utils.Optional(overrides...) {
		return nil, errors.ErrAlreadyExists(cpi.KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", c.GetName(), c.repo.arch.GetVersion()))
	}
	return newComponentVersionAccess(c, version, false)
}

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionContainer struct {
	impl support.ComponentVersionAccessImpl

	comp *ComponentAccessImpl

	descriptor *compdesc.ComponentDescriptor
}

var _ support.ComponentVersionContainer = (*ComponentVersionContainer)(nil)

func newComponentVersionAccess(comp *ComponentAccessImpl, version string, persistent bool) (cpi.ComponentVersionAccess, error) {
	c, err := newComponentVersionContainer(comp)
	if err != nil {
		return nil, err
	}
	impl, err := support.NewComponentVersionAccessImpl(comp.GetName(), version, c, true, persistent, !compositionmodeattr.Get(comp.GetContext()))
	if err != nil {
		c.Close()
		return nil, err
	}
	return cpi.NewComponentVersionAccess(impl), nil
}

func newComponentVersionContainer(comp *ComponentAccessImpl) (*ComponentVersionContainer, error) {
	return &ComponentVersionContainer{
		comp:       comp,
		descriptor: comp.repo.arch.GetDescriptor(),
	}, nil
}

func (c *ComponentVersionContainer) SetImplementation(impl support.ComponentVersionAccessImpl) {
	c.impl = impl
}

func (c *ComponentVersionContainer) GetParentViewManager() cpi.ComponentAccessViewManager {
	return c.comp
}

func (c *ComponentVersionContainer) Close() error {
	return nil
}

func (c *ComponentVersionContainer) GetContext() cpi.Context {
	return c.comp.GetContext()
}

func (c *ComponentVersionContainer) Repository() cpi.Repository {
	return c.comp.repo.arch.nonref
}

func (c *ComponentVersionContainer) IsReadOnly() bool {
	return c.comp.repo.arch.IsReadOnly()
}

func (c *ComponentVersionContainer) Update() error {
	desc := c.comp.repo.arch.GetDescriptor()
	*desc = *c.descriptor.Copy()
	return c.comp.repo.arch.container.Update()
}

func (c *ComponentVersionContainer) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.descriptor
}

func (c *ComponentVersionContainer) GetBlobData(name string) (cpi.DataAccess, error) {
	return c.comp.repo.arch.container.GetBlobData(name)
}

func (c *ComponentVersionContainer) GetStorageContext(cv cpi.ComponentVersionAccess) cpi.StorageContext {
	return ocmhdlr.New(c.Repository(), cv, &BlobSink{c.comp.repo.arch.container.base}, Type)
}

func (c *ComponentVersionContainer) AddBlobFor(storagectx cpi.StorageContext, blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	err := c.comp.repo.arch.container.base.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	return localblob.New(common.DigestToFileName(blob.Digest()), refName, blob.MimeType(), global), nil
}

func (c *ComponentVersionContainer) AccessMethod(a cpi.AccessSpec) (cpi.AccessMethod, error) {
	if a.GetKind() == localblob.Type || a.GetKind() == localfsblob.Type {
		accessSpec, err := c.GetContext().AccessSpecForSpec(a)
		if err != nil {
			return nil, err
		}
		return newLocalFilesystemBlobAccessMethod(accessSpec.(*localblob.AccessSpec), c), nil
	}
	return nil, errors.ErrNotSupported(errors.KIND_ACCESSMETHOD, a.GetType(), "component archive")
}

func (c *ComponentVersionContainer) GetInexpensiveContentVersionIdentity(a cpi.AccessSpec) string {
	if a.GetKind() == localblob.Type || a.GetKind() == localfsblob.Type {
		accessSpec, err := c.GetContext().AccessSpecForSpec(a)
		if err != nil {
			return ""
		}
		digest, _ := accessio.Digest(newLocalFilesystemBlobAccessMethod(accessSpec.(*localblob.AccessSpec), c))
		return digest.String()
	}
	return ""
}
