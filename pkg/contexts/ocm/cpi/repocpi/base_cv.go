// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
)

// here, we define the common implementation agnostic parts
// for component version objects referred to by a ComponentVersionView.

// ComponentVersionAccessImpl is the provider implementation
// interface for component versions.
type ComponentVersionAccessImpl interface {
	GetContext() cpi.Context
	SetImplementation(base ComponentVersionAccessBase)
	GetParentViewManager() ComponentAccessViewManager

	Repository() cpi.Repository

	IsReadOnly() bool

	GetDescriptor() *compdesc.ComponentDescriptor

	AccessMethod(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error)
	GetInexpensiveContentVersionIdentity(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) string

	Update() error

	BlobContainer
	io.Closer
}

type _componentVersionAccessImplBase = resource.ResourceImplBase[cpi.ComponentVersionAccess]

// componentVersionAccessBase is the counterpart to views, all views
// created by Dup calls use this base object to work on.
// Besides some functionality covered by view objects these base objects
// implement provider-agnostic parts of the ComponentVersionAccess API.
type componentVersionAccessBase struct {
	*_componentVersionAccessImplBase
	ctx     cpi.Context
	name    string
	version string

	blobcache BlobCache

	lazy           bool
	directAccess   bool
	persistent     bool
	discardChanges bool

	impl ComponentVersionAccessImpl
}

var _ ComponentVersionAccessBase = (*componentVersionAccessBase)(nil)

func newComponentVersionAccessBase(name, version string, impl ComponentVersionAccessImpl, lazy, persistent, direct bool, closer ...io.Closer) (ComponentVersionAccessBase, error) {
	base, err := resource.NewResourceImplBase[cpi.ComponentVersionAccess](impl.GetParentViewManager(), closer...)
	if err != nil {
		return nil, err
	}
	b := &componentVersionAccessBase{
		_componentVersionAccessImplBase: base,
		ctx:                             impl.GetContext(),
		name:                            name,
		version:                         version,
		blobcache:                       NewBlobCache(),
		lazy:                            lazy,
		persistent:                      persistent,
		directAccess:                    direct,
		impl:                            impl,
	}
	impl.SetImplementation(b)
	return b, nil
}

func GetComponentVersionImpl[T ComponentVersionAccessImpl](cv cpi.ComponentVersionAccess) (T, error) {
	var _nil T

	impl, err := GetComponentVersionAccessBase(cv)
	if err != nil {
		return _nil, err
	}
	if mine, ok := impl.(*componentVersionAccessBase); ok {
		cont, ok := mine.impl.(T)
		if ok {
			return cont, nil
		}
		return _nil, errors.Newf("non-matching component version implementation %T", mine.impl)
	}
	return _nil, errors.Newf("non-matching component version implementation %T", impl)
}

func (b *componentVersionAccessBase) Close() error {
	list := errors.ErrListf("closing component version %s", common.VersionedElementKey(b))
	refmgmt.AllocLog.Trace("closing component version base", "name", common.VersionedElementKey(b))
	list.Add(b.impl.Close())
	list.Add(b._componentVersionAccessImplBase.Close())
	list.Add(b.blobcache.Clear())
	refmgmt.AllocLog.Trace("closed component version base", "name", common.VersionedElementKey(b))
	return list.Result()
}

func (b *componentVersionAccessBase) GetContext() cpi.Context {
	return b.ctx
}

func (b *componentVersionAccessBase) GetName() string {
	return b.name
}

func (b *componentVersionAccessBase) GetVersion() string {
	return b.version
}

func (b *componentVersionAccessBase) GetBlobCache() BlobCache {
	return b.blobcache
}

func (b *componentVersionAccessBase) EnablePersistence() bool {
	if b.discardChanges {
		return false
	}
	b.persistent = true
	b.GetStorageContext()
	return true
}

func (b *componentVersionAccessBase) IsPersistent() bool {
	return b.persistent
}

func (b *componentVersionAccessBase) UseDirectAccess() bool {
	return b.directAccess
}

func (b *componentVersionAccessBase) DiscardChanges() {
	b.discardChanges = true
}

func (b *componentVersionAccessBase) Repository() cpi.Repository {
	return b.impl.Repository()
}

func (b *componentVersionAccessBase) IsReadOnly() bool {
	return b.impl.IsReadOnly()
}

////////////////////////////////////////////////////////////////////////////////
// with access to actual view

func (b *componentVersionAccessBase) AccessMethod(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error) {
	return b.impl.AccessMethod(acc, cv)
}

func (b *componentVersionAccessBase) GetInexpensiveContentVersionIdentity(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) string {
	return b.impl.GetInexpensiveContentVersionIdentity(acc, cv)
}

func (b *componentVersionAccessBase) GetDescriptor() *compdesc.ComponentDescriptor {
	return b.impl.GetDescriptor()
}

func (b *componentVersionAccessBase) GetStorageContext() cpi.StorageContext {
	return b.impl.GetStorageContext()
}

func (b *componentVersionAccessBase) AddBlobFor(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	return b.impl.AddBlobFor(blob, refName, global)
}

func (b *componentVersionAccessBase) ShouldUpdate(final bool) bool {
	if b.discardChanges {
		return false
	}
	if final {
		return b.persistent
	}
	return !b.lazy && b.directAccess && b.persistent
}

func (b *componentVersionAccessBase) Update(final bool) error {
	if b.ShouldUpdate(final) {
		return b.impl.Update()
	}
	return nil
}
