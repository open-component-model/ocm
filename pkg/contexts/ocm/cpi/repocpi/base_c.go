// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/blobaccess"
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

	cvbase, err := GetComponentVersionAccessBase(cv)
	if err != nil {
		return err
	}

	forcestore := c.IsOwned(cv)
	if !forcestore {
		eff, err := c.NewVersion(cv.GetVersion(), optionutils.AsValue(opts.Overwrite))
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

		d := eff.GetDescriptor()
		*d = *cv.GetDescriptor().Copy()

		err = c.setupLocalBlobs("resource", cv, cvbase, d.Resources, &opts.BlobUploadOptions)
		if err == nil {
			err = c.setupLocalBlobs("source", cv, cvbase, d.Sources, &opts.BlobUploadOptions)
		}
		if err != nil {
			return err
		}
	}
	cvbase.EnablePersistence()
	err = cvbase.Update(!cvbase.UseDirectAccess())
	return err
}

func (c *componentAccessBase) setupLocalBlobs(kind string, src cpi.ComponentVersionAccess, tgtbase ComponentVersionAccessBase, it compdesc.ArtifactAccessor, opts *cpi.BlobUploadOptions) (ferr error) {
	ctx := src.GetContext()
	// transfer all local blobs
	prov := func(spec cpi.AccessSpec) (blob blobaccess.BlobAccess, ref string, global cpi.AccessSpec, err error) {
		if spec.IsLocal(ctx) {
			m, err := spec.AccessMethod(src)
			if err != nil {
				return nil, "", nil, err
			}
			return m.AsBlobAccess(), cpi.ReferenceHint(spec, src), cpi.GlobalAccess(spec, tgtbase.GetContext()), nil
		}
		return nil, "", nil, nil
	}

	return tgtbase.(*componentVersionAccessBase).setupLocalBlobs(kind, prov, it, false, opts)
}
