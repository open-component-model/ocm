// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"fmt"
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
	"github.com/open-component-model/ocm/pkg/utils"
)

type _componentAccessView interface {
	resource.ResourceViewInt[cpi.ComponentAccess] // here you have to redeclare
}

type ComponentAccessViewManager = resource.ViewManager[cpi.ComponentAccess] // here you have to use an alias

type ComponentAccessProxy interface {
	resource.ResourceImplementation[cpi.ComponentAccess]

	GetContext() cpi.Context
	IsReadOnly() bool
	GetName() string

	IsOwned(access cpi.ComponentVersionAccess) bool

	ListVersions() ([]string, error)
	LookupVersion(version string) (cpi.ComponentVersionAccess, error)
	HasVersion(vers string) (bool, error)
	NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error)

	Close() error
	AddVersion(cv cpi.ComponentVersionAccess, opts *cpi.AddVersionOptions) (ferr error)
}

type componentAccessView struct {
	_componentAccessView
	proxy ComponentAccessProxy
}

var (
	_ cpi.ComponentAccess = (*componentAccessView)(nil)
	_ utils.Unwrappable   = (*componentAccessView)(nil)
)

func GetComponentAccessProxy(n cpi.ComponentAccess) (ComponentAccessProxy, error) {
	if v, ok := n.(*componentAccessView); ok {
		return v.proxy, nil
	}
	return nil, errors.ErrNotSupported("component base type", fmt.Sprintf("%T", n))
}

func GetComponentAccessImplementation(n cpi.ComponentAccess) (ComponentAccessImpl, error) {
	if v, ok := n.(*componentAccessView); ok {
		if b, ok := v.proxy.(*componentAccessProxy); ok {
			return b.impl, nil
		}
		return nil, errors.ErrNotSupported("component base type", fmt.Sprintf("%T", v.proxy))
	}
	return nil, errors.ErrNotSupported("component implementation type", fmt.Sprintf("%T", n))
}

func componentAccessViewCreator(i ComponentAccessProxy, v resource.CloserView, d ComponentAccessViewManager) cpi.ComponentAccess {
	return &componentAccessView{
		_componentAccessView: resource.NewView[cpi.ComponentAccess](v, d),
		proxy:                i,
	}
}

func NewComponentAccess(impl ComponentAccessImpl, kind string, main bool, closer ...io.Closer) (cpi.ComponentAccess, error) {
	proxy, err := newComponentAccessProxy(impl, closer...)
	if err != nil {
		return nil, errors.Join(err, impl.Close())
	}
	if kind == "" {
		kind = "component"
	}
	cv := resource.NewResource[cpi.ComponentAccess](proxy, componentAccessViewCreator, fmt.Sprintf("%s %s", kind, impl.GetName()), main)
	return cv, nil
}

func (c *componentAccessView) Unwrap() interface{} {
	return c.proxy
}

func (c *componentAccessView) GetContext() cpi.Context {
	return c.proxy.GetContext()
}

func (c *componentAccessView) GetName() string {
	return c.proxy.GetName()
}

func (c *componentAccessView) ListVersions() (list []string, err error) {
	err = c.Execute(func() error {
		list, err = c.proxy.ListVersions()
		return err
	})
	return list, err
}

func (c *componentAccessView) LookupVersion(version string) (acc cpi.ComponentVersionAccess, err error) {
	err = c.Execute(func() error {
		acc, err = c.proxy.LookupVersion(version)
		return err
	})
	return acc, err
}

func (c *componentAccessView) AddVersion(acc cpi.ComponentVersionAccess, overwrite ...bool) error {
	if acc.GetName() != c.GetName() {
		return errors.ErrInvalid("component name", acc.GetName())
	}

	return c.Execute(func() error {
		return c.proxy.AddVersion(acc, cpi.NewAddVersionOptions(cpi.Overwrite(utils.Optional(overwrite...))))
	})
}

func (c *componentAccessView) AddVersionOpt(acc cpi.ComponentVersionAccess, opts ...cpi.AddVersionOption) error {
	if acc.GetName() != c.GetName() {
		return errors.ErrInvalid("component name", acc.GetName())
	}
	return c.Execute(func() error {
		return c.proxy.AddVersion(acc, cpi.NewAddVersionOptions(opts...))
	})
}

func (c *componentAccessView) NewVersion(version string, overrides ...bool) (acc cpi.ComponentVersionAccess, err error) {
	err = c.Execute(func() error {
		if c.proxy.IsReadOnly() {
			return accessio.ErrReadOnly
		}
		acc, err = c.proxy.NewVersion(version, overrides...)
		return err
	})
	return acc, err
}

func (c *componentAccessView) HasVersion(vers string) (ok bool, err error) {
	err = c.Execute(func() error {
		ok, err = c.proxy.HasVersion(vers)
		return err
	})
	return ok, err
}
