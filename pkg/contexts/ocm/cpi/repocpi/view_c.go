// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"fmt"
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
	"github.com/open-component-model/ocm/pkg/utils"
)

type _componentAccessView interface {
	resource.ResourceViewInt[cpi.ComponentAccess] // here you have to redeclare
}

type ComponentAccessViewManager = resource.ViewManager[cpi.ComponentAccess] // here you have to use an alias

type ComponentAccessBase interface {
	resource.ResourceImplementation[cpi.ComponentAccess]
	internal.ComponentAccessImpl

	IsReadOnly() bool
	GetName() string

	IsOwned(access cpi.ComponentVersionAccess) bool

	AddVersion(cv cpi.ComponentVersionAccess) error
}

type componentAccessView struct {
	_componentAccessView
	base ComponentAccessBase
}

var (
	_ cpi.ComponentAccess = (*componentAccessView)(nil)
	_ utils.Unwrappable   = (*componentAccessView)(nil)
)

func GetComponentAccessBase(n cpi.ComponentAccess) (ComponentAccessBase, error) {
	if v, ok := n.(*componentAccessView); ok {
		return v.base, nil
	}
	return nil, errors.ErrNotSupported("component base type", fmt.Sprintf("%T", n))
}

func GetComponentAccessImplementation(n cpi.ComponentAccess) (ComponentAccessImpl, error) {
	if v, ok := n.(*componentAccessView); ok {
		if b, ok := v.base.(*componentAccessBase); ok {
			return b.impl, nil
		}
		return nil, errors.ErrNotSupported("component base type", fmt.Sprintf("%T", v.base))
	}
	return nil, errors.ErrNotSupported("component implementation type", fmt.Sprintf("%T", n))
}

func componentAccessViewCreator(i ComponentAccessBase, v resource.CloserView, d ComponentAccessViewManager) cpi.ComponentAccess {
	return &componentAccessView{
		_componentAccessView: resource.NewView[cpi.ComponentAccess](v, d),
		base:                 i,
	}
}

func NewComponentAccess(impl ComponentAccessImpl, kind string, closer ...io.Closer) (cpi.ComponentAccess, error) {
	base, err := newComponentAccessImplBase(impl, closer...)
	if err != nil {
		return nil, errors.Join(err, impl.Close())
	}
	if kind == "" {
		kind = "component"
	}
	cv := resource.NewResource[cpi.ComponentAccess](base, componentAccessViewCreator, fmt.Sprintf("%s %s", kind, impl.GetName()), true)
	return cv, nil
}

func (c *componentAccessView) Unwrap() interface{} {
	return c.base
}

func (c *componentAccessView) GetContext() cpi.Context {
	return c.base.GetContext()
}

func (c *componentAccessView) GetName() string {
	return c.base.GetName()
}

func (c *componentAccessView) ListVersions() (list []string, err error) {
	err = c.Execute(func() error {
		list, err = c.base.ListVersions()
		return err
	})
	return list, err
}

func (c *componentAccessView) LookupVersion(version string) (acc cpi.ComponentVersionAccess, err error) {
	err = c.Execute(func() error {
		acc, err = c.base.LookupVersion(version)
		return err
	})
	return acc, err
}

func (c *componentAccessView) AddVersion(acc cpi.ComponentVersionAccess, overrides ...bool) error {
	if acc.GetName() != c.GetName() {
		return errors.ErrInvalid("component name", acc.GetName())
	}
	return c.Execute(func() error {
		return c.addVersion(acc, overrides...)
	})
}

func (c *componentAccessView) addVersion(acc cpi.ComponentVersionAccess, overrides ...bool) (ferr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&ferr)

	ctx := acc.GetContext()

	impl, err := GetComponentVersionAccessBase(acc)
	if err != nil {
		return err
	}

	var (
		d   *compdesc.ComponentDescriptor
		sel func(cpi.AccessSpec) bool
		eff cpi.ComponentVersionAccess
	)

	opts := cpi.NewBlobUploadOptions()

	forcestore := c.base.IsOwned(acc)
	if !forcestore {
		// transfer all local blobs into a new owned version.
		sel = func(spec cpi.AccessSpec) bool { return spec.IsLocal(ctx) }

		eff, err = c.base.NewVersion(acc.GetVersion(), overrides...)
		if err != nil {
			return err
		}
		finalize.With(func() error {
			return eff.Close()
		})
		impl, err = GetComponentVersionAccessBase(eff)
		if err != nil {
			return err
		}

		d = eff.GetDescriptor()
		*d = *acc.GetDescriptor().Copy()
	} else {
		// transfer composition blobs into local blobs
		opts.UseNoDefaultIfNotSet = true
		opts.BlobHandlerProvider = nil
		sel = compose.Is
		d = acc.GetDescriptor()
		eff = acc
	}

	err = setupLocalBlobs(ctx, "resource", acc, nil, impl, d.Resources, sel, forcestore, opts)
	if err == nil {
		err = setupLocalBlobs(ctx, "source", acc, nil, impl, d.Sources, sel, forcestore, opts)
	}
	if err != nil {
		return err
	}

	return c.base.AddVersion(eff)
}

func (c *componentAccessView) NewVersion(version string, overrides ...bool) (acc cpi.ComponentVersionAccess, err error) {
	err = c.Execute(func() error {
		if c.base.IsReadOnly() {
			return accessio.ErrReadOnly
		}
		acc, err = c.base.NewVersion(version, overrides...)
		return err
	})
	return acc, err
}

func (c *componentAccessView) HasVersion(vers string) (ok bool, err error) {
	err = c.Execute(func() error {
		ok, err = c.base.HasVersion(vers)
		return err
	})
	return ok, err
}
