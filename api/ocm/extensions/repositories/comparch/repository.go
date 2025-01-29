package comparch

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localfsblob"
	ocmhdlr "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/ocm"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/errkind"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/refmgmt"
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type RepositoryImpl struct {
	lock   sync.RWMutex
	bridge repocpi.RepositoryBridge
	arch   *ComponentArchive
	nonref cpi.Repository
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var _ repocpi.RepositoryImpl = (*RepositoryImpl)(nil)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func NewRepository(ctxp cpi.ContextProvider, s *RepositorySpec) (cpi.Repository, error) {
	ctx := datacontext.InternalContextRef(ctxp.OCMContext())
	if s.GetPathFileSystem() == nil {
		s.SetPathFileSystem(vfsattr.Get(ctx))
	}
	a, err := Open(ctx, s.AccessMode, s.FilePath, 0o700, s)
	if err != nil {
		return nil, err
	}
	return a.AsRepository(), nil
}

func newRepository(a *ComponentArchive) (main, nonref cpi.Repository) {
	// close main cv abstraction on repository close -------v
	impl := &RepositoryImpl{
		arch: a,
	}
	r := repocpi.NewRepository(impl, "comparch")
	return r, impl.nonref
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) Close() error {
	return r.arch.container.Close()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) IsReadOnly() bool {
	return r.arch.IsReadOnly()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) SetReadOnly() {
	r.arch.SetReadOnly()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) SetBridge(base repocpi.RepositoryBridge) {
	r.bridge = base
	r.nonref = repocpi.NewNoneRefRepositoryView(base)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) GetContext() cpi.Context {
	return r.arch.GetContext()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) ComponentLister() cpi.ComponentLister {
	return r
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
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

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
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

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
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

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) Get() *ComponentArchive {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch != nil {
		return r.arch
	}
	return nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) GetSpecification() cpi.RepositorySpec {
	return r.arch.spec
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) ExistsComponentVersion(name string, ref string) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return false, accessio.ErrClosed
	}
	return r.arch.GetName() == name && r.arch.GetVersion() == ref, nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (r *RepositoryImpl) LookupComponent(name string) (*repocpi.ComponentAccessInfo, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.arch == nil {
		return nil, accessio.ErrClosed
	}
	if r.arch.GetName() != name {
		return nil, errors.ErrNotFound(errkind.KIND_COMPONENT, name, Type)
	}
	return newComponentAccess(r)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type ComponentAccessImpl struct {
	base repocpi.ComponentAccessBridge
	repo *RepositoryImpl
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var _ repocpi.ComponentAccessImpl = (*ComponentAccessImpl)(nil)

func newComponentAccess(r *RepositoryImpl) (*repocpi.ComponentAccessInfo, error) {
	impl := &ComponentAccessImpl{
		repo: r,
	}
	return &repocpi.ComponentAccessInfo{impl, "component archive", true}, nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) Close() error {
	return nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) SetBridge(base repocpi.ComponentAccessBridge) {
	c.base = base
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) GetParentBridge() repocpi.RepositoryViewManager {
	return c.repo.bridge
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) GetContext() cpi.Context {
	return c.repo.GetContext()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) GetName() string {
	return c.repo.arch.GetName()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) IsReadOnly() bool {
	return c.repo.arch.IsReadOnly()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) ListVersions() ([]string, error) {
	return []string{c.repo.arch.GetVersion()}, nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) HasVersion(vers string) (bool, error) {
	return vers == c.repo.arch.GetVersion(), nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) LookupVersion(version string) (*repocpi.ComponentVersionAccessInfo, error) {
	if version != c.repo.arch.GetVersion() {
		return nil, errors.ErrNotFound(cpi.KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", c.GetName(), c.repo.arch.GetVersion()))
	}
	return newComponentVersionAccess(c, version, false)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentAccessImpl) NewVersion(version string, overrides ...bool) (*repocpi.ComponentVersionAccessInfo, error) {
	if version != c.repo.arch.GetVersion() {
		return nil, errors.ErrNotSupported(cpi.KIND_COMPONENTVERSION, version, fmt.Sprintf("component archive %s:%s", c.GetName(), c.repo.arch.GetVersion()))
	}
	if !utils.Optional(overrides...) {
		return nil, errors.ErrAlreadyExists(cpi.KIND_COMPONENTVERSION, fmt.Sprintf("%s:%s", c.GetName(), c.repo.arch.GetVersion()))
	}
	return newComponentVersionAccess(c, version, false)
}

////////////////////////////////////////////////////////////////////////////////

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type ComponentVersionContainer struct {
	impl repocpi.ComponentVersionAccessBridge

	comp *ComponentAccessImpl

	descriptor *compdesc.ComponentDescriptor
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var _ repocpi.ComponentVersionAccessImpl = (*ComponentVersionContainer)(nil)

func newComponentVersionAccess(comp *ComponentAccessImpl, version string, persistent bool) (*repocpi.ComponentVersionAccessInfo, error) {
	c, err := newComponentVersionContainer(comp)
	if err != nil {
		return nil, err
	}
	return &repocpi.ComponentVersionAccessInfo{c, true, persistent}, nil
}

func newComponentVersionContainer(comp *ComponentAccessImpl) (*ComponentVersionContainer, error) {
	return &ComponentVersionContainer{
		comp:       comp,
		descriptor: comp.repo.arch.GetDescriptor(),
	}, nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) SetBridge(impl repocpi.ComponentVersionAccessBridge) {
	c.impl = impl
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) GetParentBridge() repocpi.ComponentAccessBridge {
	return c.comp.base
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) Close() error {
	return nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) GetContext() cpi.Context {
	return c.comp.GetContext()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) Repository() cpi.Repository {
	return c.comp.repo.arch.nonref
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) IsReadOnly() bool {
	return c.comp.repo.arch.IsReadOnly()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) SetReadOnly() {
	c.comp.repo.arch.SetReadOnly()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) Update() (bool, error) {
	desc := c.comp.repo.arch.GetDescriptor()
	*desc = *c.descriptor.Copy()
	return c.comp.repo.arch.container.Update()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) SetDescriptor(cd *compdesc.ComponentDescriptor) (bool, error) {
	*c.descriptor = *cd
	return c.Update()
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.descriptor
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) GetBlob(name string) (cpi.DataAccess, error) {
	return c.comp.repo.arch.container.GetBlob(name)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) GetStorageContext() cpi.StorageContext {
	return ocmhdlr.New(c.Repository(), c.comp.GetName(), &BlobSink{c.comp.repo.arch.container.fsacc}, Type)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) AddBlob(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	err := c.comp.repo.arch.container.fsacc.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	return localblob.New(common.DigestToFileName(blob.Digest()), refName, blob.MimeType(), global), nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (c *ComponentVersionContainer) AccessMethod(a cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error) {
	if a.GetKind() == localblob.Type || a.GetKind() == localfsblob.Type {
		accessSpec, err := c.GetContext().AccessSpecForSpec(a)
		if err != nil {
			return nil, err
		}
		return newLocalFilesystemBlobAccessMethod(accessSpec.(*localblob.AccessSpec), c, cv)
	}
	return nil, errors.ErrNotSupported(errkind.KIND_ACCESSMETHOD, a.GetType(), "component archive")
}
