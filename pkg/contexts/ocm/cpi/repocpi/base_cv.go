// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/keepblobattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/optionutils"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
	"github.com/open-component-model/ocm/pkg/utils"
)

// here, we define the common implementation agnostic parts
// for component version objects referred to by a ComponentVersionView.

// ComponentVersionAccessImpl is the provider implementation
// interface for component versions.
type ComponentVersionAccessImpl interface {
	GetContext() cpi.Context
	SetBase(base ComponentVersionAccessBase)
	GetParentBase() ComponentAccessBase

	Repository() cpi.Repository

	IsReadOnly() bool

	GetDescriptor() *compdesc.ComponentDescriptor
	SetDescriptor(*compdesc.ComponentDescriptor) error

	AccessMethod(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error)
	GetInexpensiveContentVersionIdentity(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) string

	BlobContainer
	io.Closer
}

type ComponentVersionAccessImplSupport struct {
	Base ComponentVersionAccessBase
}

func (b *ComponentVersionAccessImplSupport) SetBase(base ComponentVersionAccessBase) {
	b.Base = base
}

type _componentVersionAccessImplBase = resource.ResourceImplBase[cpi.ComponentVersionAccess]

// componentVersionAccessBase is the counterpart to views, all views
// created by Dup calls use this base object to work on.
// Besides some functionality covered by view objects these base objects
// implement provider-agnostic parts of the ComponentVersionAccess API.
type componentVersionAccessBase struct {
	lock sync.Mutex

	*_componentVersionAccessImplBase
	ctx     cpi.Context
	name    string
	version string

	descriptor *compdesc.ComponentDescriptor
	blobcache  BlobCache

	lazy           bool
	directAccess   bool
	persistent     bool
	discardChanges bool

	impl ComponentVersionAccessImpl
}

var _ ComponentVersionAccessBase = (*componentVersionAccessBase)(nil)

func newComponentVersionAccessBase(name, version string, impl ComponentVersionAccessImpl, lazy, persistent, direct bool, closer ...io.Closer) (ComponentVersionAccessBase, error) {
	base, err := resource.NewResourceImplBase[cpi.ComponentVersionAccess, cpi.ComponentAccess](impl.GetParentBase(), closer...)
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
	impl.SetBase(b)
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

func (b *componentVersionAccessBase) GetImplementation() ComponentVersionAccessImpl {
	return b.impl
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
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.descriptor == nil {
		b.descriptor = b.impl.GetDescriptor()
	}
	return b.descriptor
}

func (b *componentVersionAccessBase) GetStorageContext() cpi.StorageContext {
	return b.impl.GetStorageContext()
}

func (b *componentVersionAccessBase) ShouldUpdate(final bool) bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.shouldUpdate(final)
}

func (b *componentVersionAccessBase) shouldUpdate(final bool) bool {
	if b.discardChanges {
		return false
	}
	if final {
		return b.persistent
	}
	return !b.lazy && b.directAccess && b.persistent
}

func (b *componentVersionAccessBase) Update(final bool) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.shouldUpdate(final) {
		return b.impl.SetDescriptor(b.descriptor.Copy())
	}
	return nil
}

func (c *componentVersionAccessBase) AddBlob(blob cpi.BlobAccess, artType, refName string, global cpi.AccessSpec, final bool, opts *cpi.BlobUploadOptions) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	if c.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	blob, err := blob.Dup()
	if err != nil {
		return nil, errors.Wrapf(err, "invalid blob access")
	}
	defer blob.Close()
	err = utils.ValidateObject(blob)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid blob access")
	}

	storagectx := c.GetStorageContext()
	ctx := c.GetContext()

	// handle foreign blob upload
	var prov cpi.BlobHandlerProvider
	if opts.BlobHandlerProvider != nil {
		prov = opts.BlobHandlerProvider
	} else {
		if !optionutils.AsValue(opts.UseNoDefaultIfNotSet) {
			prov = internal.BlobHandlerProviderForRegistry(ctx.BlobHandlers())
		} else {
			//nolint: staticcheck // yes
			// use no blob uploader
		}
	}
	if prov != nil {
		h := prov.LookupHandler(storagectx, artType, blob.MimeType())
		if h != nil {
			acc, err := h.StoreBlob(blob, artType, refName, nil, storagectx)
			if err != nil {
				return nil, err
			}
			if acc != nil {
				if !keepblobattr.Get(ctx) || acc.IsLocal(ctx) {
					return acc, nil
				}
				global = acc
			}
		}
	}

	var acc cpi.AccessSpec

	if final || c.UseDirectAccess() {
		acc, err = c.impl.AddBlob(blob, refName, global)
		if err != nil {
			return nil, err
		}
	} else {
		// use local composition access to be added to the repository with AddVersion.
		acc = compose.New(refName, blob.MimeType(), global)
	}
	return c.cacheLocalBlob(acc, blob)
}

func (c *componentVersionAccessBase) cacheLocalBlob(acc cpi.AccessSpec, blob cpi.BlobAccess) (cpi.AccessSpec, error) {
	key, err := json.Marshal(acc)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal access spec")
	}
	// local blobs might not be accessible from the underlying
	// repository implementation if the component version is not
	// finally added (for example ghcr.io as OCI repository).
	// Therefore, we keep a copy of the blob access for further usage.

	// if a local blob is uploader and the access method is replaced
	// we have to handle the case that the technical upload repo
	// is the same as the storage backend of the OCM repository, which
	// might have been configured with local credentials, which were
	// reused by the uploader.
	// The access spec is independent of the actual repo, so it does
	// not have access to those credentials. Therefore, we have to
	// keep the original blob for further usage, also.
	err = c.GetBlobCache().AddBlobFor(string(key), blob)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
