// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
)

// ComponentAccessImpl is the provider implementation
// interface for component versions.
type ComponentAccessImpl interface {
	SetImplementation(base ComponentAccessBase)
	GetParentViewManager() RepositoryViewManager

	GetContext() cpi.Context
	GetName() string
	IsReadOnly() bool

	ListVersions() ([]string, error)
	LookupVersion(version string) (cpi.ComponentVersionAccess, error)
	HasVersion(vers string) (bool, error)
	NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error)
	AddVersion(cv cpi.ComponentVersionAccess) error

	io.Closer
}

type _componentAccessImplBase = resource.ResourceImplBase[cpi.ComponentAccess]

type componentAccessBase struct {
	*_componentAccessImplBase
	ctx  cpi.Context
	name string
	impl ComponentAccessImpl
}

func newComponentAccessImplBase(impl ComponentAccessImpl, closer ...io.Closer) (ComponentAccessBase, error) {
	base, err := resource.NewResourceImplBase[cpi.ComponentAccess](impl.GetParentViewManager(), closer...)
	if err != nil {
		return nil, err
	}
	b := &componentAccessBase{
		_componentAccessImplBase: base,
		ctx:                      impl.GetContext(),
		name:                     impl.GetName(),
		impl:                     impl,
	}
	impl.SetImplementation(b)
	return b, nil
}

func (b *componentAccessBase) Close() error {
	list := errors.ErrListf("closing component %s", b.name)
	refmgmt.AllocLog.Trace("closing component base", "name", b.name)
	list.Add(b.impl.Close())
	list.Add(b._componentAccessImplBase.Close())
	refmgmt.AllocLog.Trace("closed component base", "name", b.name)
	return list.Result()
}

func (b *componentAccessBase) GetContext() cpi.Context {
	return b.ctx
}

func (b *componentAccessBase) GetName() string {
	return b.name
}

func (c *componentAccessBase) IsOwned(cv cpi.ComponentVersionAccess) bool {
	base, err := GetComponentVersionAccessBase(cv)
	if err != nil {
		return false
	}

	impl := base.(*componentVersionAccessBase).impl
	cvcompmgr := impl.GetParentViewManager()
	mymgr := c._componentAccessImplBase
	return mymgr == cvcompmgr
}

func (c *componentAccessBase) AddVersion(cv cpi.ComponentVersionAccess) error {
	return c.impl.AddVersion(cv)
}

func (b *componentAccessBase) ListVersions() ([]string, error) {
	return b.impl.ListVersions()
}

func (b *componentAccessBase) LookupVersion(version string) (internal.ComponentVersionAccess, error) {
	return b.impl.LookupVersion(version)
}

func (b *componentAccessBase) HasVersion(vers string) (bool, error) {
	return b.impl.HasVersion(vers)
}

func (b *componentAccessBase) NewVersion(version string, overrides ...bool) (internal.ComponentVersionAccess, error) {
	return b.impl.NewVersion(version, overrides...)
}

func (b *componentAccessBase) IsReadOnly() bool {
	return b.impl.IsReadOnly()
}
