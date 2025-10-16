package virtual

import (
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localfsblob"
	ocmhdlr "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/ocm"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/errkind"
	"ocm.software/ocm/api/utils/refmgmt"
)

// newComponentVersionAccess creates a component access for the artifact access, if this fails the artifact access is closed.
func newComponentVersionAccess(comp *componentAccessImpl, version string, persistent bool) (*repocpi.ComponentVersionAccessInfo, error) {
	access, err := comp.repo.access.GetComponentVersion(comp.GetName(), version)
	if err != nil {
		return nil, err
	}
	c, err := newComponentVersionContainer(comp, version, access)
	if err != nil {
		return nil, err
	}
	return &repocpi.ComponentVersionAccessInfo{c, true, persistent}, nil
}

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionContainer struct {
	bridge repocpi.ComponentVersionAccessBridge

	comp    *componentAccessImpl
	version string
	access  VersionAccess
}

var _ repocpi.ComponentVersionAccessImpl = (*ComponentVersionContainer)(nil)

func newComponentVersionContainer(comp *componentAccessImpl, version string, access VersionAccess) (*ComponentVersionContainer, error) {
	return &ComponentVersionContainer{
		comp:    comp,
		version: version,
		access:  access,
	}, nil
}

func (c *ComponentVersionContainer) SetBridge(base repocpi.ComponentVersionAccessBridge) {
	c.bridge = base
}

func (c *ComponentVersionContainer) GetParentBridge() repocpi.ComponentAccessBridge {
	return c.comp.bridge
}

func (c *ComponentVersionContainer) Close() error {
	if c.access == nil {
		return accessio.ErrClosed
	}
	a := c.access
	c.access = nil
	return a.Close()
}

func (c *ComponentVersionContainer) Check() error {
	if c.version != c.GetDescriptor().Version {
		return errors.ErrInvalid("component version", c.GetDescriptor().Version)
	}
	if c.comp.name != c.GetDescriptor().Name {
		return errors.ErrInvalid("component name", c.GetDescriptor().Name)
	}
	return nil
}

func (c *ComponentVersionContainer) Repository() cpi.Repository {
	return c.comp.repo.nonref
}

func (c *ComponentVersionContainer) GetContext() cpi.Context {
	return c.comp.GetContext()
}

func (c *ComponentVersionContainer) IsReadOnly() bool {
	return c.access.IsReadOnly()
}

func (c *ComponentVersionContainer) SetReadOnly() {
	c.access.SetReadOnly()
}

func (c *ComponentVersionContainer) IsClosed() bool {
	return c.access == nil
}

func (c *ComponentVersionContainer) AccessMethod(a cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error) {
	accessSpec, err := c.comp.GetContext().AccessSpecForSpec(a)
	if err != nil {
		return nil, err
	}

	switch a.GetKind() { // to be extended
	case localfsblob.Type:
		fallthrough
	case localblob.Type:
		blob, err := c.access.GetBlob(accessSpec.(*localblob.AccessSpec).LocalReference)
		if err != nil {
			return nil, err
		}

		return accspeccpi.AccessMethodForImplementation(newLocalBlobAccessMethod(accessSpec.(*localblob.AccessSpec), blob))
	}

	return nil, errors.ErrNotSupported(errkind.KIND_ACCESSMETHOD, a.GetType(), "virtual registry")
}

func (c *ComponentVersionContainer) Update() (bool, error) {
	return c.access.Update()
}

func (c *ComponentVersionContainer) SetDescriptor(cd *compdesc.ComponentDescriptor) (bool, error) {
	cur := c.access.GetDescriptor()
	*cur = *cd
	return c.access.Update()
}

func (c *ComponentVersionContainer) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.access.GetDescriptor()
}

func (c *ComponentVersionContainer) GetBlob(name string) (cpi.DataAccess, error) {
	return c.access.GetBlob(name)
}

func (c *ComponentVersionContainer) GetStorageContext() cpi.StorageContext {
	return ocmhdlr.New(c.Repository(), c.comp.GetName(), c.access, Type, c.access)
}

func (c *ComponentVersionContainer) AddBlob(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	if c.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}

	ref, err := c.access.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	return localblob.New(ref, refName, blob.MimeType(), global), nil
}
