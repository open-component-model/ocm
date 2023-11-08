// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package virtual

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	ocmhdlr "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/handlers/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/repocpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

// newComponentVersionAccess creates a component access for the artifact access, if this fails the artifact acess is closed.
func newComponentVersionAccess(comp *componentAccessImpl, version string, persistent bool) (cpi.ComponentVersionAccess, error) {
	access, err := comp.repo.access.GetComponentVersion(comp.GetName(), version)
	if err != nil {
		return nil, err
	}
	c, err := newComponentVersionContainer(comp, version, access)
	if err != nil {
		return nil, err
	}
	return repocpi.NewComponentVersionAccess(comp.GetName(), version, c, true, persistent, !compositionmodeattr.Get(comp.GetContext()))
}

// //////////////////////////////////////////////////////////////////////////////

type ComponentVersionContainer struct {
	base repocpi.ComponentVersionAccessBase

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

func (c *ComponentVersionContainer) SetImplementation(base repocpi.ComponentVersionAccessBase) {
	c.base = base
}

func (c *ComponentVersionContainer) GetParentViewManager() repocpi.ComponentAccessViewManager {
	return c.comp.base
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

	return nil, errors.ErrNotSupported(errors.KIND_ACCESSMETHOD, a.GetType(), "virtual registry")
}

func (c *ComponentVersionContainer) GetInexpensiveContentVersionIdentity(a cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) string {
	accessSpec, err := c.comp.GetContext().AccessSpecForSpec(a)
	if err != nil {
		return ""
	}

	switch a.GetKind() { // to be extended
	case localfsblob.Type:
		fallthrough
	case localblob.Type:
		return c.access.GetInexpensiveContentVersionIdentity(accessSpec)
	}

	return ""
}

func (c *ComponentVersionContainer) Update() error {
	return c.access.Update()
}

func (c *ComponentVersionContainer) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.access.GetDescriptor()
}

func (c *ComponentVersionContainer) GetBlobData(name string) (cpi.DataAccess, error) {
	return c.access.GetBlob(name)
}

func (c *ComponentVersionContainer) GetStorageContext() cpi.StorageContext {
	return ocmhdlr.New(c.Repository(), c.comp.GetName(), c.access, Type, c.access)
}

func (c *ComponentVersionContainer) AddBlobFor(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
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
