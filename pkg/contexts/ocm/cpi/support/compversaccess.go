// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package support

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

type _ComponentVersionAccessImplBase = cpi.ComponentVersionAccessImplBase

type ComponentVersionAccessImpl interface {
	cpi.ComponentVersionAccessImpl
	EnablePersistence() bool
}

type componentVersionAccessImpl struct {
	*_ComponentVersionAccessImplBase
	lazy           bool
	directAccess   bool
	persistent     bool
	discardChanges bool
	base           ComponentVersionContainer
}

var _ ComponentVersionAccessImpl = (*componentVersionAccessImpl)(nil)

func GetComponentVersionContainer[T ComponentVersionContainer](cv cpi.ComponentVersionAccess) (T, error) {
	var _nil T

	impl, err := cpi.GetComponentVersionAccessImplementation(cv)
	if err != nil {
		return _nil, err
	}
	if mine, ok := impl.(*componentVersionAccessImpl); ok {
		cont, ok := mine.base.(T)
		if ok {
			return cont, nil
		}
		return _nil, errors.Newf("non-matching component version implementation %T", mine.base)
	}
	return _nil, errors.Newf("non-matching component version implementation %T", impl)
}

func NewComponentVersionAccessImpl(name, version string, container ComponentVersionContainer, lazy, persistent, direct bool) (cpi.ComponentVersionAccessImpl, error) {
	base, err := cpi.NewComponentVersionAccessImplBase(container.GetContext(), name, version, container.GetParentViewManager())
	if err != nil {
		return nil, err
	}
	impl := &componentVersionAccessImpl{
		_ComponentVersionAccessImplBase: base,
		lazy:                            lazy,
		persistent:                      persistent,
		directAccess:                    direct,
		base:                            container,
	}
	container.SetImplementation(impl)
	return impl, nil
}

func (a *componentVersionAccessImpl) EnablePersistence() bool {
	if a.discardChanges {
		return false
	}
	a.persistent = true
	a.GetStorageContext()
	return true
}

func (a *componentVersionAccessImpl) IsPersistent() bool {
	return a.persistent
}

func (d *componentVersionAccessImpl) UseDirectAccess() bool {
	return d.directAccess
}

func (a *componentVersionAccessImpl) DiscardChanges() {
	a.discardChanges = true
}

func (a *componentVersionAccessImpl) Close() error {
	list := errors.ErrListf("closing component version access %s/%s", a.GetName(), a.GetVersion())
	return list.Add(a.base.Close(), a._ComponentVersionAccessImplBase.Close()).Result()
}

func (a *componentVersionAccessImpl) Repository() cpi.Repository {
	return a.base.Repository()
}

func (a *componentVersionAccessImpl) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

////////////////////////////////////////////////////////////////////////////////
// with access to actual view

func (a *componentVersionAccessImpl) AccessMethod(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error) {
	return a.base.AccessMethod(acc, cv)
}

func (a *componentVersionAccessImpl) GetInexpensiveContentVersionIdentity(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) string {
	return a.base.GetInexpensiveContentVersionIdentity(acc, cv)
}

func (a *componentVersionAccessImpl) GetDescriptor() *compdesc.ComponentDescriptor {
	return a.base.GetDescriptor()
}

func (a *componentVersionAccessImpl) GetStorageContext() cpi.StorageContext {
	return a.base.GetStorageContext()
}

func (a *componentVersionAccessImpl) AddBlobFor(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	return a.base.AddBlobFor(blob, refName, global)
}

func (a *componentVersionAccessImpl) ShouldUpdate(final bool) bool {
	if a.discardChanges {
		return false
	}
	if final {
		return a.persistent
	}
	return !a.lazy && a.directAccess && a.persistent
}

func (a *componentVersionAccessImpl) Update(final bool) error {
	if a.ShouldUpdate(final) {
		return a.base.Update()
	}
	return nil
}
