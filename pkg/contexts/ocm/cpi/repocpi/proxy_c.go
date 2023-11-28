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
	SetProxy(proxy ComponentAccessProxy)
	GetParentProxy() RepositoryViewManager

	GetContext() cpi.Context
	GetName() string
	IsReadOnly() bool

	ListVersions() ([]string, error)
	HasVersion(vers string) (bool, error)
	LookupVersion(version string) (*ComponentVersionAccessInfo, error)
	NewVersion(version string, overrides ...bool) (*ComponentVersionAccessInfo, error)

	io.Closer
}

type _componentAccessProxyBase = resource.ResourceImplBase[cpi.ComponentAccess]

type componentAccessProxy struct {
	*_componentAccessProxyBase
	ctx  cpi.Context
	name string
	impl ComponentAccessImpl
}

func newComponentAccessProxy(impl ComponentAccessImpl, closer ...io.Closer) (ComponentAccessProxy, error) {
	base, err := resource.NewResourceImplBase[cpi.ComponentAccess, cpi.Repository](impl.GetParentProxy(), closer...)
	if err != nil {
		return nil, err
	}
	b := &componentAccessProxy{
		_componentAccessProxyBase: base,
		ctx:                       impl.GetContext(),
		name:                      impl.GetName(),
		impl:                      impl,
	}
	impl.SetProxy(b)
	return b, nil
}

func (b *componentAccessProxy) Close() error {
	list := errors.ErrListf("closing component %s", b.name)
	refmgmt.AllocLog.Trace("closing component proxy", "name", b.name)
	list.Add(b.impl.Close())
	list.Add(b._componentAccessProxyBase.Close())
	refmgmt.AllocLog.Trace("closed component proxy", "name", b.name)
	return list.Result()
}

func (b *componentAccessProxy) GetContext() cpi.Context {
	return b.ctx
}

func (b *componentAccessProxy) GetName() string {
	return b.name
}

func (b *componentAccessProxy) IsReadOnly() bool {
	return b.impl.IsReadOnly()
}

func (c *componentAccessProxy) IsOwned(cv cpi.ComponentVersionAccess) bool {
	proxy, err := GetComponentVersionAccessProxy(cv)
	if err != nil {
		return false
	}

	impl := proxy.(*componentVersionAccessProxy).impl
	cvcompmgr := impl.GetParentProxy()
	return c == cvcompmgr
}

func (b *componentAccessProxy) ListVersions() ([]string, error) {
	return b.impl.ListVersions()
}

func (b *componentAccessProxy) LookupVersion(version string) (cpi.ComponentVersionAccess, error) {
	i, err := b.impl.LookupVersion(version)
	if err != nil {
		return nil, err
	}
	if i == nil || i.Impl == nil {
		return nil, errors.ErrInvalid("component implementation behaviour", "LookupVersion")
	}
	return NewComponentVersionAccess(b.GetName(), version, i.Impl, i.Lazy, i.Persistent, !compositionmodeattr.Get(b.GetContext()))
}

func (b *componentAccessProxy) HasVersion(vers string) (bool, error) {
	return b.impl.HasVersion(vers)
}

func (b *componentAccessProxy) NewVersion(version string, overrides ...bool) (cpi.ComponentVersionAccess, error) {
	i, err := b.impl.NewVersion(version, overrides...)
	if err != nil {
		return nil, err
	}
	if i == nil || i.Impl == nil {
		return nil, errors.ErrInvalid("component implementation behaviour", "NewVersion")
	}
	return NewComponentVersionAccess(b.GetName(), version, i.Impl, i.Lazy, false, !compositionmodeattr.Get(b.GetContext()))
}

func (c *componentAccessProxy) AddVersion(cv cpi.ComponentVersionAccess, opts *cpi.AddVersionOptions) (ferr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&ferr)

	cvproxy, err := GetComponentVersionAccessProxy(cv)
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
		cvproxy, err = GetComponentVersionAccessProxy(eff)
		if err != nil {
			return err
		}

		d := eff.GetDescriptor()
		*d = *cv.GetDescriptor().Copy()

		err = c.setupLocalBlobs("resource", cv, cvproxy, d.Resources, &opts.BlobUploadOptions)
		if err == nil {
			err = c.setupLocalBlobs("source", cv, cvproxy, d.Sources, &opts.BlobUploadOptions)
		}
		if err != nil {
			return err
		}
	}
	cvproxy.EnablePersistence()
	err = cvproxy.Update(!cvproxy.UseDirectAccess())
	return err
}

func (c *componentAccessProxy) setupLocalBlobs(kind string, src cpi.ComponentVersionAccess, tgtproxy ComponentVersionAccessProxy, it compdesc.ArtifactAccessor, opts *cpi.BlobUploadOptions) (ferr error) {
	ctx := src.GetContext()
	// transfer all local blobs
	prov := func(spec cpi.AccessSpec) (blob blobaccess.BlobAccess, ref string, global cpi.AccessSpec, err error) {
		if spec.IsLocal(ctx) {
			m, err := spec.AccessMethod(src)
			if err != nil {
				return nil, "", nil, err
			}
			return m.AsBlobAccess(), cpi.ReferenceHint(spec, src), cpi.GlobalAccess(spec, tgtproxy.GetContext()), nil
		}
		return nil, "", nil, nil
	}

	return tgtproxy.(*componentVersionAccessProxy).setupLocalBlobs(kind, prov, it, false, opts)
}
