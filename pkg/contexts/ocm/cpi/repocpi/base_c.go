// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/optionutils"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
)

type ComponentVersionAccessInfo struct {
	Impl       ComponentVersionAccessImpl
	Lazy       bool
	Persistent bool
}

// ComponentAccessImpl is the provider implementation
// interface for component versions.
type ComponentAccessImpl interface {
	SetBase(base ComponentAccessBase)
	GetParentBase() RepositoryViewManager

	GetContext() cpi.Context
	GetName() string
	IsReadOnly() bool

	ListVersions() ([]string, error)
	HasVersion(vers string) (bool, error)
	LookupVersion(version string) (*ComponentVersionAccessInfo, error)
	NewVersion(version string, overrides ...bool) (*ComponentVersionAccessInfo, error)

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
	base, err := resource.NewResourceImplBase[cpi.ComponentAccess, cpi.Repository](impl.GetParentBase(), closer...)
	if err != nil {
		return nil, err
	}
	b := &componentAccessBase{
		_componentAccessImplBase: base,
		ctx:                      impl.GetContext(),
		name:                     impl.GetName(),
		impl:                     impl,
	}
	impl.SetBase(b)
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

func (b *componentAccessBase) IsReadOnly() bool {
	return b.impl.IsReadOnly()
}

func (c *componentAccessBase) IsOwned(cv cpi.ComponentVersionAccess) bool {
	base, err := GetComponentVersionAccessBase(cv)
	if err != nil {
		return false
	}

	impl := base.(*componentVersionAccessBase).impl
	cvcompmgr := impl.GetParentBase()
	return c == cvcompmgr
}

func (b *componentAccessBase) ListVersions() ([]string, error) {
	return b.impl.ListVersions()
}

func (b *componentAccessBase) LookupVersion(version string) (cpi.ComponentVersionAccess, error) {
	i, err := b.impl.LookupVersion(version)
	if err != nil {
		return nil, err
	}
	if i == nil || i.Impl == nil {
		return nil, errors.ErrInvalid("component implementation behaviour", "LookupVersion")
	}
	return NewComponentVersionAccess(b.GetName(), version, i.Impl, i.Lazy, i.Persistent, !compositionmodeattr.Get(b.GetContext()))
}

func (b *componentAccessBase) HasVersion(vers string) (bool, error) {
	return b.impl.HasVersion(vers)
}

func (b *componentAccessBase) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	i, err := b.impl.NewVersion(version, overrides...)
	if err != nil {
		return nil, err
	}
	if i == nil || i.Impl == nil {
		return nil, errors.ErrInvalid("component implementation behaviour", "NewVersion")
	}
	return NewComponentVersionAccess(b.GetName(), version, i.Impl, i.Lazy, false, !compositionmodeattr.Get(b.GetContext()))
}

func (c *componentAccessBase) AddVersion(cv cpi.ComponentVersionAccess, opts *cpi.AddVersionOptions) (ferr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&ferr)

	ctx := c.GetContext()
	cvbase, err := GetComponentVersionAccessBase(cv)
	if err != nil {
		return err
	}

	var (
		d   *compdesc.ComponentDescriptor
		sel func(cpi.AccessSpec) bool
		eff cpi.ComponentVersionAccess
	)

	forcestore := c.IsOwned(cv)
	if !forcestore {
		// transfer all local blobs into a new owned version.
		sel = func(spec cpi.AccessSpec) bool { return spec.IsLocal(ctx) }

		eff, err = c.NewVersion(cv.GetVersion(), optionutils.AsValue(opts.Overwrite))
		if err != nil {
			return err
		}
		finalize.With(func() error {
			return eff.Close()
		})
		cvbase, err = GetComponentVersionAccessBase(eff)
		if err != nil {
			return err
		}

		d = eff.GetDescriptor()
		*d = *cv.GetDescriptor().Copy()
	} else {
		// transfer composition blobs into local blobs
		opts.UseNoDefaultIfNotSet = optionutils.PointerTo(true)
		opts.BlobHandlerProvider = nil
		sel = compose.Is
		d = cv.GetDescriptor()
		eff = cv
	}

	err = setupLocalBlobs(ctx, "resource", cv, nil, cvbase, d.Resources, sel, forcestore, &opts.BlobUploadOptions)
	if err == nil {
		err = setupLocalBlobs(ctx, "source", cv, nil, cvbase, d.Sources, sel, forcestore, &opts.BlobUploadOptions)
	}
	if err != nil {
		return err
	}

	cvbase.EnablePersistence()
	return cvbase.Update(!cvbase.UseDirectAccess())
}
