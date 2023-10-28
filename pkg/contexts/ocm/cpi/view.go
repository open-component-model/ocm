// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/keepblobattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
	"github.com/open-component-model/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/utils/selector"
)

// View objects are the user facing generic implementations of the context interfaces.
// They are responsible to handle the reference counting and use
// shared implementations objects for th concrete type-specific implementations.
// Additionally, they are used to implement interface functionality which is
// common to all implementations and NOT dependent on the backend system technology.

var (
	ErrClosed      = resource.ErrClosed
	ErrTempVersion = fmt.Errorf("temporary component version cannot be updated")
)

////////////////////////////////////////////////////////////////////////////////

type _RepositoryView interface {
	resource.ResourceViewInt[Repository] // here you have to redeclare
}

type RepositoryViewManager = resource.ViewManager[Repository] // here you have to use an alias

type RepositoryImpl interface {
	resource.ResourceImplementation[Repository]
	internal.RepositoryImpl
}

type _RepositoryImplBase = resource.ResourceImplBase[Repository]

type RepositoryImplBase struct {
	_RepositoryImplBase
	ctx Context
}

func (b *RepositoryImplBase) GetContext() Context {
	return b.ctx
}

func NewRepositoryImplBase(ctx Context, closer ...io.Closer) *RepositoryImplBase {
	base, _ := resource.NewResourceImplBase[Repository, io.Closer](nil, closer...)
	return &RepositoryImplBase{
		_RepositoryImplBase: *base,
		ctx:                 ctx,
	}
}

type repositoryView struct {
	_RepositoryView
	impl RepositoryImpl
}

var (
	_ Repository                           = (*repositoryView)(nil)
	_ credentials.ConsumerIdentityProvider = (*repositoryView)(nil)
)

func GetRepositoryImplementation(n Repository) (RepositoryImpl, error) {
	if v, ok := n.(*repositoryView); ok {
		return v.impl, nil
	}
	return nil, errors.ErrNotSupported("repository implementation type", fmt.Sprintf("%T", n))
}

func repositoryViewCreator(i RepositoryImpl, v resource.CloserView, d RepositoryViewManager) Repository {
	return &repositoryView{
		_RepositoryView: resource.NewView[Repository](v, d),
		impl:            i,
	}
}

// NewNoneRefRepositoryView provides a repository reflecting the state of the
// view manager without holding an additional reference.
func NewNoneRefRepositoryView(i RepositoryImpl) Repository {
	return &repositoryView{
		_RepositoryView: resource.NewView[Repository](resource.NewNonRefView[Repository](i), i),
		impl:            i,
	}
}

func NewRepository(impl RepositoryImpl, name ...string) Repository {
	return resource.NewResource[Repository](impl, repositoryViewCreator, utils.OptionalDefaulted("OCM repo", name...), true)
}

func (r *repositoryView) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	return credentials.GetProvidedConsumerId(r.impl, uctx...)
}

func (r *repositoryView) GetIdentityMatcher() string {
	return credentials.GetProvidedIdentityMatcher(r.impl)
}

func (r *repositoryView) GetSpecification() RepositorySpec {
	return r.impl.GetSpecification()
}

func (r *repositoryView) GetContext() Context {
	return r.impl.GetContext()
}

func (r *repositoryView) ComponentLister() ComponentLister {
	return r.impl.ComponentLister()
}

func (r *repositoryView) ExistsComponentVersion(name string, version string) (ok bool, err error) {
	err = r.Execute(func() error {
		ok, err = r.impl.ExistsComponentVersion(name, version)
		return err
	})
	return ok, err
}

func (r *repositoryView) LookupComponentVersion(name string, version string) (acc ComponentVersionAccess, err error) {
	err = r.Execute(func() error {
		acc, err = r.impl.LookupComponentVersion(name, version)
		return err
	})
	return acc, err
}

func (r *repositoryView) LookupComponent(name string) (acc ComponentAccess, err error) {
	err = r.Execute(func() error {
		acc, err = r.impl.LookupComponent(name)
		return err
	})
	return acc, err
}

func (r *repositoryView) NewComponentVersion(comp, vers string, overrides ...bool) (ComponentVersionAccess, error) {
	c, err := refmgmt.ToLazy(r.LookupComponent(comp))
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c.NewVersion(vers, overrides...)
}

func (r *repositoryView) AddComponentVersion(cv ComponentVersionAccess, overrides ...bool) error {
	c, err := refmgmt.ToLazy(r.LookupComponent(cv.GetName()))
	if err != nil {
		return err
	}
	defer c.Close()

	return c.AddVersion(cv, overrides...)
}

////////////////////////////////////////////////////////////////////////////////

type _ComponentAccessView interface {
	resource.ResourceViewInt[ComponentAccess] // here you have to redeclare
}

type ComponentAccessViewManager = resource.ViewManager[ComponentAccess] // here you have to use an alias

type ComponentAccessImpl interface {
	resource.ResourceImplementation[ComponentAccess]
	internal.ComponentAccessImpl

	IsReadOnly() bool
	GetName() string

	IsOwned(access ComponentVersionAccess) bool

	AddVersion(cv ComponentVersionAccess) error
}

type _ComponentAccessImplBase = resource.ResourceImplBase[ComponentAccess]

type ComponentAccessImplBase struct {
	*_ComponentAccessImplBase
	ctx  Context
	name string
}

func NewComponentAccessImplBase(ctx Context, name string, repo RepositoryViewManager, closer ...io.Closer) (*ComponentAccessImplBase, error) {
	base, err := resource.NewResourceImplBase[ComponentAccess](repo, closer...)
	if err != nil {
		return nil, err
	}
	return &ComponentAccessImplBase{
		_ComponentAccessImplBase: base,
		ctx:                      ctx,
		name:                     name,
	}, nil
}

func (b *ComponentAccessImplBase) GetContext() Context {
	return b.ctx
}

func (b *ComponentAccessImplBase) GetName() string {
	return b.name
}

type componentAccessView struct {
	_ComponentAccessView
	impl ComponentAccessImpl
}

var _ ComponentAccess = (*componentAccessView)(nil)

func GetComponentAccessImplementation(n ComponentAccess) (ComponentAccessImpl, error) {
	if v, ok := n.(*componentAccessView); ok {
		return v.impl, nil
	}
	return nil, errors.ErrNotSupported("component implementation type", fmt.Sprintf("%T", n))
}

func componentAccessViewCreator(i ComponentAccessImpl, v resource.CloserView, d ComponentAccessViewManager) ComponentAccess {
	return &componentAccessView{
		_ComponentAccessView: resource.NewView[ComponentAccess](v, d),
		impl:                 i,
	}
}

func NewComponentAccess(impl ComponentAccessImpl, kind ...string) ComponentAccess {
	return resource.NewResource[ComponentAccess](impl, componentAccessViewCreator, fmt.Sprintf("%s %s", utils.OptionalDefaulted("component", kind...), impl.GetName()), true)
}

func (c *componentAccessView) GetContext() Context {
	return c.impl.GetContext()
}

func (c *componentAccessView) GetName() string {
	return c.impl.GetName()
}

func (c *componentAccessView) ListVersions() (list []string, err error) {
	err = c.Execute(func() error {
		list, err = c.impl.ListVersions()
		return err
	})
	return list, err
}

func (c *componentAccessView) LookupVersion(version string) (acc ComponentVersionAccess, err error) {
	err = c.Execute(func() error {
		acc, err = c.impl.LookupVersion(version)
		return err
	})
	return acc, err
}

func (c *componentAccessView) AddVersion(acc ComponentVersionAccess, overrides ...bool) error {
	if acc.GetName() != c.GetName() {
		return errors.ErrInvalid("component name", acc.GetName())
	}
	return c.Execute(func() error {
		return c.addVersion(acc, overrides...)
	})
}

func (c *componentAccessView) addVersion(acc ComponentVersionAccess, overrides ...bool) (ferr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&ferr)

	ctx := acc.GetContext()

	impl, err := GetComponentVersionAccessImplementation(acc)
	if err != nil {
		return err
	}

	var (
		d   *compdesc.ComponentDescriptor
		sel func(AccessSpec) bool
		eff ComponentVersionAccess
	)

	if !c.impl.IsOwned(acc) {
		// transfer all local blobs into a new owned version.
		sel = func(spec AccessSpec) bool { return spec.IsLocal(ctx) }

		eff, err = c.impl.NewVersion(acc.GetVersion(), overrides...)
		if err != nil {
			return err
		}
		finalize.With(func() error {
			return eff.Close()
		})
		impl, err = GetComponentVersionAccessImplementation(eff)
		if err != nil {
			return err
		}

		d = eff.GetDescriptor()
		*d = *acc.GetDescriptor().Copy()
	} else {
		// transfer composition blobs into local blobs
		sel = compose.Is
		d = acc.GetDescriptor()
		eff = acc
	}

	err = setupLocalBobs(ctx, "resource", acc, eff, nil, impl, d.Resources, sel)
	if err == nil {
		err = setupLocalBobs(ctx, "source", acc, eff, nil, impl, d.Sources, sel)
	}
	if err != nil {
		return err
	}

	return c.impl.AddVersion(eff)
}

func setupLocalBobs(ctx Context, kind string, src, tgt ComponentVersionAccess, accprov func(AccessSpec) (AccessMethod, error), tgtimpl ComponentVersionAccessImpl, it compdesc.ArtifactAccessor, sel func(AccessSpec) bool) (ferr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&ferr)

	for i := 0; i < it.Len(); i++ {
		nested := finalize.Nested()
		a := it.GetArtifact(i)
		spec, err := ctx.AccessSpecForSpec(a.GetAccess())
		if err != nil {
			return errors.Wrapf(err, "%s %d", kind, i)
		}
		if sel(spec) {
			blob, err := blobAccessForLocalAccessSpec(spec, src, accprov)
			if err != nil {
				return errors.Wrapf(err, "%s %d", kind, i)
			}
			nested.Close(blob)
			effspec, err := addBlob(tgtimpl, tgt, a.GetType(), ReferenceHint(spec, src), blob, GlobalAccess(spec, ctx))
			if err != nil {
				return errors.Wrapf(err, "cannot store %s %d", kind, i)
			}
			a.SetAccess(effspec)
		}
		err = nested.Finalize()
		if err != nil {
			return errors.Wrapf(err, "%s %d", kind, i)
		}
	}
	return nil
}

func blobAccessForLocalAccessSpec(spec AccessSpec, cv ComponentVersionAccess, accprov func(AccessSpec) (AccessMethod, error)) (blobaccess.BlobAccess, error) {
	var m AccessMethod
	var err error
	if accprov != nil {
		m, err = accprov(spec)
	} else {
		m, err = spec.AccessMethod(cv)
	}
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	v := AccessMethodAsView(m)
	defer v.Close()
	return BlobAccessForAccessMethod(v)
}

func (c *componentAccessView) NewVersion(version string, overrides ...bool) (acc ComponentVersionAccess, err error) {
	err = c.Execute(func() error {
		if c.impl.IsReadOnly() {
			return accessio.ErrReadOnly
		}
		acc, err = c.impl.NewVersion(version, overrides...)
		return err
	})
	return acc, err
}

func (c *componentAccessView) HasVersion(vers string) (ok bool, err error) {
	err = c.Execute(func() error {
		ok, err = c.impl.HasVersion(vers)
		return err
	})
	return ok, err
}

////////////////////////////////////////////////////////////////////////////////

type _ComponentVersionAccessView interface {
	resource.ResourceViewInt[ComponentVersionAccess]
}

type ComponentVersionAccessViewManager = resource.ViewManager[ComponentVersionAccess]

type ComponentVersionAccessImpl interface {
	resource.ResourceImplementation[ComponentVersionAccess]
	common.VersionedElement
	io.Closer

	GetContext() Context
	Repository() Repository

	DiscardChanges()
	IsPersistent() bool

	GetDescriptor() *compdesc.ComponentDescriptor

	AccessMethod(ComponentVersionAccess, AccessSpec) (AccessMethod, error)

	GetInexpensiveContentVersionIdentity(ComponentVersionAccess, AccessSpec) string

	// GetStorageContext creates a storage context for blobs
	// that is used to feed blob handlers for specific blob storage methods.
	// If no handler accepts the blob, the AddBlobFor method will
	// be used to store the blob
	GetStorageContext(cv ComponentVersionAccess) StorageContext

	// AddBlobFor stores a local blob together with the component and
	// potentially provides a global reference.
	// The resulting access information (global and local) is provided as
	// an access method specification usable in a component descriptor.
	// This is the direct technical storage, without caring about any handler.
	AddBlobFor(storagectx StorageContext, blob BlobAccess, refName string, global AccessSpec) (AccessSpec, error)

	IsReadOnly() bool

	// ShouldUpdate checks, whether an update is indicated
	// by the state of object, considering persistence, lazy, discard
	// and update mode state
	ShouldUpdate(final bool) bool

	// GetBlobCache retieves the blob cache used to store preliminary
	// blob accesses for freshly generated local access specs not directly
	// usable until a component version is finally added to the repository.
	GetBlobCache() BlobCache

	// UseDirectAccess returns true if composition should be directly
	// forwarded to the repository backend.,
	UseDirectAccess() bool

	// Update persists the current state of the component version to the
	// underlying repository backend.
	Update(final bool) error
}

type (
	BlobCacheEntry = blobaccess.BlobAccess
	BlobCacheKey   = interface{}
)

type BlobCache interface {
	// AddBlobFor stores blobs for added blobs not yet accessible
	// by generated access method until version is finally added.
	AddBlobFor(acc BlobCacheKey, blob BlobCacheEntry) error

	// GetBlobFor retrieves the original blob access for
	// a given access specification.
	GetBlobFor(acc BlobCacheKey) BlobCacheEntry

	RemoveBlobFor(acc BlobCacheKey)
	Clear() error
}

type blobCache struct {
	lock      sync.Mutex
	blobcache map[BlobCacheKey]BlobCacheEntry
}

func NewBlobCache() BlobCache {
	return &blobCache{
		blobcache: map[BlobCacheKey]BlobCacheEntry{},
	}
}

func (c *blobCache) RemoveBlobFor(acc BlobCacheKey) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if b := c.blobcache[acc]; b != nil {
		b.Close()
		delete(c.blobcache, acc)
	}
}

func (c *blobCache) AddBlobFor(acc BlobCacheKey, blob BlobCacheEntry) error {
	if s, ok := acc.(string); ok && s == "" {
		return errors.ErrInvalid("blob key")
	}
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.blobcache[acc] == nil {
		l, err := blob.Dup()
		if err != nil {
			return err
		}
		c.blobcache[acc] = l
	}
	return nil
}

func (c *blobCache) GetBlobFor(acc BlobCacheKey) BlobCacheEntry {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.blobcache[acc]
}

func (c *blobCache) Clear() error {
	list := errors.ErrList()
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, b := range c.blobcache {
		list.Add(b.Close())
	}
	c.blobcache = map[BlobCacheKey]BlobCacheEntry{}
	return list.Result()
}

type _ComponentVersionAccessImplBase = resource.ResourceImplBase[ComponentVersionAccess]

type ComponentVersionAccessImplBase struct {
	*_ComponentVersionAccessImplBase
	ctx     Context
	name    string
	version string

	blobcache BlobCache
}

func NewComponentVersionAccessImplBase(ctx Context, name, version string, repo ComponentAccessViewManager, closer ...io.Closer) (*ComponentVersionAccessImplBase, error) {
	base, err := resource.NewResourceImplBase[ComponentVersionAccess](repo, closer...)
	if err != nil {
		return nil, err
	}
	return &ComponentVersionAccessImplBase{
		_ComponentVersionAccessImplBase: base,
		ctx:                             ctx,
		name:                            name,
		version:                         version,
		blobcache:                       NewBlobCache(),
	}, nil
}

func (b *ComponentVersionAccessImplBase) Close() error {
	list := errors.ErrListf("closing %s", common.VersionedElementKey(b))
	list.Add(b._ComponentVersionAccessImplBase.Close())
	list.Add(b.blobcache.Clear())
	return list.Result()
}

func (b *ComponentVersionAccessImplBase) GetContext() Context {
	return b.ctx
}

func (b *ComponentVersionAccessImplBase) GetName() string {
	return b.name
}

func (b *ComponentVersionAccessImplBase) GetVersion() string {
	return b.version
}

func (b *ComponentVersionAccessImplBase) GetBlobCache() BlobCache {
	return b.blobcache
}

type componentVersionAccessView struct {
	_ComponentVersionAccessView
	impl ComponentVersionAccessImpl
}

var _ ComponentVersionAccess = (*componentVersionAccessView)(nil)

func GetComponentVersionAccessImplementation(n ComponentVersionAccess) (ComponentVersionAccessImpl, error) {
	if v, ok := n.(*componentVersionAccessView); ok {
		return v.impl, nil
	}
	return nil, errors.ErrNotSupported("component version implementation type", fmt.Sprintf("%T", n))
}

func artifactAccessViewCreator(i ComponentVersionAccessImpl, v resource.CloserView, d resource.ViewManager[ComponentVersionAccess]) ComponentVersionAccess {
	return &componentVersionAccessView{
		_ComponentVersionAccessView: resource.NewView[ComponentVersionAccess](v, d),
		impl:                        i,
	}
}

func NewComponentVersionAccess(impl ComponentVersionAccessImpl) ComponentVersionAccess {
	return resource.NewResource[ComponentVersionAccess](impl, artifactAccessViewCreator, fmt.Sprintf("component version  %s/%s", impl.GetName(), impl.GetVersion()), true)
}

func (c *componentVersionAccessView) Close() error {
	err := c.Execute(func() error {
		// executed under local lock, if refcount is one, I'm the last user.
		if c.impl.RefCount() == 1 {
			// prepare artifact access for final close in
			// direct access mode.
			if !compositionmodeattr.Get(c.GetContext()) {
				err := c.update(true)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return c._ComponentVersionAccessView.Close()
}

func (c *componentVersionAccessView) Repository() Repository {
	return c.impl.Repository()
}

func (c *componentVersionAccessView) GetContext() internal.Context {
	return c.impl.GetContext()
}

func (c *componentVersionAccessView) GetName() string {
	return c.impl.GetName()
}

func (c *componentVersionAccessView) GetVersion() string {
	return c.impl.GetVersion()
}

func (c *componentVersionAccessView) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.impl.GetDescriptor()
}

func (c *componentVersionAccessView) GetProvider() *compdesc.Provider {
	return c.GetDescriptor().Provider.Copy()
}

func (c *componentVersionAccessView) SetProvider(p *compdesc.Provider) error {
	return c.Execute(func() error {
		c.GetDescriptor().Provider = *p.Copy()
		return nil
	})
}

func (c *componentVersionAccessView) AccessMethod(spec AccessSpec) (meth AccessMethod, err error) {
	spec, err = c.GetContext().AccessSpecForSpec(spec)
	if err != nil {
		return nil, err
	}
	err = c.Execute(func() error {
		var err error
		meth, err = c.accessMethod(spec)
		return err
	})
	return meth, err
}

func (c *componentVersionAccessView) accessMethod(spec AccessSpec) (meth AccessMethod, err error) {
	switch {
	case compose.Is(spec):
		cspec, ok := spec.(*compose.AccessSpec)
		if !ok {
			return nil, fmt.Errorf("invalid implementation (%T) for access method compose", spec)
		}
		blob := c.getLocalBlob(cspec)
		if blob == nil {
			return nil, errors.ErrUnknown(blobaccess.KIND_BLOB, cspec.Id, common.VersionedElementKey(c).String())
		}
		meth, err = compose.NewMethod(cspec, blob)
	case !spec.IsLocal(c.GetContext()):
		meth, err = spec.AccessMethod(c)
	default:
		meth, err = c.impl.AccessMethod(c, spec)
		if err == nil {
			if blob := c.getLocalBlob(spec); blob != nil {
				meth, err = newFakeMethod(meth, blob)
			}
		}
	}
	return meth, err
}

func (c *componentVersionAccessView) GetInexpensiveContentVersionIdentity(spec AccessSpec) string {
	var err error

	spec, err = c.GetContext().AccessSpecForSpec(spec)
	if err != nil {
		return ""
	}

	var id string
	_ = c.Execute(func() error {
		id = c.getInexpensiveContentVersionIdentity(spec)
		return nil
	})
	return id
}

func (c *componentVersionAccessView) getInexpensiveContentVersionIdentity(spec AccessSpec) string {
	switch {
	case compose.Is(spec):
		fallthrough
	case !spec.IsLocal(c.GetContext()):
		// fall back to original version
		return spec.GetInexpensiveContentVersionIdentity(c)
	default:
		return c.impl.GetInexpensiveContentVersionIdentity(c, spec)
	}
}

func (c *componentVersionAccessView) Update() error {
	return c.Execute(func() error {
		if !c.impl.IsPersistent() {
			return ErrTempVersion
		}
		return c.update(true)
	})
}

func (c *componentVersionAccessView) update(final bool) error {
	if !c.impl.ShouldUpdate(final) {
		return nil
	}

	ctx := c.GetContext()
	d := c.GetDescriptor()
	impl, err := GetComponentVersionAccessImplementation(c)
	if err != nil {
		return err
	}
	// TODO: exceute for separately lockable view
	err = setupLocalBobs(ctx, "resource", c, c, c.accessMethod, impl, d.Resources, compose.Is)
	if err == nil {
		err = setupLocalBobs(ctx, "source", c, c, c.accessMethod, impl, d.Sources, compose.Is)
	}
	if err != nil {
		return err
	}

	err = c.impl.Update(true)
	if err != nil {
		return err
	}
	return c.impl.GetBlobCache().Clear()
}

func (c *componentVersionAccessView) AddBlob(blob cpi.BlobAccess, artType, refName string, global AccessSpec, opts ...internal.BlobUploadOption) (AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	if c.impl.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	blob, err := blob.Dup()
	if err != nil {
		return nil, errors.Wrapf(err, "inavlid blob access")
	}
	defer blob.Close()
	err = utils.ValidateObject(blob)
	if err != nil {
		return nil, errors.Wrapf(err, "inavlid blob access")
	}

	eff := NewBlobUploadOptions(opts...)
	if !eff.UseNoDefaultIfNotSet && eff.BlobHandlerProvider == nil {
		eff.BlobHandlerProvider = internal.DefaultBlobHandlerProvider(c.GetContext())
	}

	var acc AccessSpec
	if c.impl.UseDirectAccess() {
		acc, err = addBlob(c.impl, c, artType, refName, blob, global)
	} else {
		// use local composition access to be added to the repository with AddVersion.
		acc = compose.New(refName, blob.MimeType(), global)
	}
	if err == nil {
		return c.cacheLocalBlob(acc, blob)
	}
	return acc, err
}

func addBlob(impl ComponentVersionAccessImpl, cv ComponentVersionAccess, artType, refName string, blob BlobAccess, global AccessSpec) (AccessSpec, error) {
	storagectx := impl.GetStorageContext(cv)
	ctx := cv.GetContext()
	h := ctx.BlobHandlers().LookupHandler(storagectx.GetImplementationRepositoryType(), artType, blob.MimeType())
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
	return impl.AddBlobFor(storagectx, blob, refName, global)
}

func (c *componentVersionAccessView) getLocalBlob(acc AccessSpec) BlobAccess {
	key, err := json.Marshal(acc)
	if err != nil {
		return nil
	}
	return c.impl.GetBlobCache().GetBlobFor(string(key))
}

func (c *componentVersionAccessView) cacheLocalBlob(acc AccessSpec, blob BlobAccess) (AccessSpec, error) {
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
	err = c.impl.GetBlobCache().AddBlobFor(string(key), blob)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (c *componentVersionAccessView) AdjustResourceAccess(meta *ResourceMeta, acc compdesc.AccessSpec, opts ...internal.ModificationOption) error {
	cd := c.GetDescriptor()
	if idx := cd.GetResourceIndex(meta); idx >= 0 {
		return c.SetResource(&cd.Resources[idx].ResourceMeta, acc, opts...)
	}
	return errors.ErrUnknown(KIND_RESOURCE, meta.GetIdentity(cd.Resources).String())
}

// SetResourceBlob adds a blob resource to the component version.
func (c *componentVersionAccessView) SetResourceBlob(meta *ResourceMeta, blob cpi.BlobAccess, refName string, global AccessSpec, opts ...internal.BlobModificationOption) error {
	Logger(c).Debug("adding resource blob", "resource", meta.Name)
	if err := utils.ValidateObject(blob); err != nil {
		return err
	}
	eff := NewBlobModificationOptions(opts...)
	acc, err := c.AddBlob(blob, meta.Type, refName, global, eff)
	if err != nil {
		return fmt.Errorf("unable to add blob (component %s:%s resource %s): %w", c.GetName(), c.GetVersion(), meta.GetName(), err)
	}

	if err := c.SetResource(meta, acc, eff, ModifyResource()); err != nil {
		return fmt.Errorf("unable to set resource: %w", err)
	}

	return nil
}

func (c *componentVersionAccessView) AdjustSourceAccess(meta *SourceMeta, acc compdesc.AccessSpec) error {
	cd := c.GetDescriptor()
	if idx := cd.GetSourceIndex(meta); idx >= 0 {
		return c.SetSource(&cd.Sources[idx].SourceMeta, acc)
	}
	return errors.ErrUnknown(KIND_RESOURCE, meta.GetIdentity(cd.Resources).String())
}

func (c *componentVersionAccessView) SetSourceBlob(meta *SourceMeta, blob BlobAccess, refName string, global AccessSpec) error {
	Logger(c).Debug("adding source blob", "source", meta.Name)
	if err := utils.ValidateObject(blob); err != nil {
		return err
	}
	acc, err := c.AddBlob(blob, meta.Type, refName, global)
	if err != nil {
		return fmt.Errorf("unable to add blob: (component %s:%s source %s): %w", c.GetName(), c.GetVersion(), meta.GetName(), err)
	}

	if err := c.SetSource(meta, acc); err != nil {
		return fmt.Errorf("unable to set source: %w", err)
	}

	return nil
}

type fakeMethod struct {
	spec  AccessSpec
	local bool
	mime  string
	blob  blobaccess.BlobAccess
}

var _ AccessMethod = (*fakeMethod)(nil)

func newFakeMethod(m AccessMethod, blob BlobAccess) (AccessMethod, error) {
	b, err := blob.Dup()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot remember blob for access method")
	}
	f := &fakeMethod{
		spec:  m.AccessSpec(),
		local: m.IsLocal(),
		mime:  m.MimeType(),
		blob:  b,
	}
	err = m.Close()
	if err != nil {
		_ = b.Close()
		return nil, errors.Wrapf(err, "closing access method")
	}
	return f, nil
}

func (f *fakeMethod) MimeType() string {
	return f.mime
}

func (f *fakeMethod) IsLocal() bool {
	return f.local
}

func (f *fakeMethod) GetKind() string {
	return f.spec.GetKind()
}

func (f *fakeMethod) AccessSpec() internal.AccessSpec {
	return f.spec
}

func (f *fakeMethod) Close() error {
	return f.blob.Close()
}

func (f *fakeMethod) Reader() (io.ReadCloser, error) {
	return f.blob.Reader()
}

func (f *fakeMethod) Get() ([]byte, error) {
	return f.blob.Get()
}

func setAccess[T any, A internal.ArtifactAccess[T]](c *componentVersionAccessView, kind string, art A,
	set func(*T, compdesc.AccessSpec) error,
	setblob func(*T, BlobAccess, string, AccessSpec) error,
) error {
	if c.impl.IsReadOnly() {
		return accessio.ErrReadOnly
	}
	meta := art.Meta()
	if meta == nil {
		return errors.Newf("no meta data provided by %s access", kind)
	}
	acc, err := art.Access()
	if err != nil && !errors.IsErrNotFoundElem(err, "", descriptor.KIND_ACCESSMETHOD) {
		return err
	}

	var (
		blob   BlobAccess
		hint   string
		global AccessSpec
	)

	if acc != nil {
		if !acc.IsLocal(c.GetContext()) {
			return set(meta, acc)
		}

		blob, err = BlobAccessForAccessSpec(acc, c)
		if err != nil && errors.IsErrNotFoundElem(err, "", blobaccess.KIND_BLOB) {
			return err
		}
		hint = ReferenceHint(acc, c)
		global = GlobalAccess(acc, c.GetContext())
	}
	if blob == nil {
		blob, err = art.BlobAccess()
		if err != nil {
			return err
		}
		defer blob.Close()
	}
	if blob == nil {
		return errors.Newf("neither access nor blob specified in %s access", kind)
	}
	if v := art.ReferenceHint(); v != "" {
		hint = v
	}
	if v := art.GlobalAccess(); v != nil {
		global = v
	}
	return setblob(meta, blob, hint, global)
}

func (c *componentVersionAccessView) SetResourceAccess(art ResourceAccess, modopts ...BlobModificationOption) error {
	return setAccess(c, "resource", art,
		func(meta *ResourceMeta, acc compdesc.AccessSpec) error {
			return c.SetResource(meta, acc, NewBlobModificationOptions(modopts...))
		},
		func(meta *ResourceMeta, blob BlobAccess, hint string, global AccessSpec) error {
			return c.SetResourceBlob(meta, blob, hint, global, modopts...)
		})
}

func (c *componentVersionAccessView) SetResource(meta *internal.ResourceMeta, acc compdesc.AccessSpec, modopts ...ModificationOption) error {
	if c.impl.IsReadOnly() {
		return accessio.ErrReadOnly
	}

	res := &compdesc.Resource{
		ResourceMeta: *meta.Copy(),
		Access:       acc,
	}

	ctx := c.impl.GetContext()
	opts := internal.NewModificationOptions(modopts...)
	CompleteModificationOptions(ctx, opts)

	spec, err := c.impl.GetContext().AccessSpecForSpec(acc)
	if err != nil {
		return err
	}

	// if the blob described by the access spec has been added
	// as local blob, just reuse the stored blob access
	// to calculate the digest to circumvent credential problems
	// for access specs generated by an uploader.
	meth, err := c.AccessMethod(spec)
	if err != nil {
		return err
	}
	if blob := c.getLocalBlob(spec); blob != nil {
		var dig digest.Digest
		if s, ok := meth.(blobaccess.DigestSource); ok {
			dig = s.Digest()
		}
		err = meth.Close()
		if err != nil {
			return errors.Wrapf(err, "clsoing shadowed method")
		}
		meth, err = NewDefaultMethodForBlobAccess(c, spec, dig, blob, spec.IsLocal(c.GetContext()))
		if err != nil {
			return err
		}
	}
	defer meth.Close()

	return c.Execute(func() error {
		var old *compdesc.Resource

		if res.Relation == metav1.LocalRelation {
			if res.Version == "" {
				res.Version = c.GetVersion()
			}
		}

		cd := c.impl.GetDescriptor()
		idx := cd.GetResourceIndex(&res.ResourceMeta)
		if idx >= 0 {
			old = &cd.Resources[idx]
		}

		if old == nil {
			if !opts.IsModifyResource() && c.impl.IsPersistent() {
				return fmt.Errorf("new resource would invalidate signature")
			}
		}

		// evaluate given digesting constraints and settings
		hashAlgo, digester, digest := c.evaluateResourceDigest(res, old, *opts)
		hasher := opts.GetHasher(hashAlgo)
		if digester.HashAlgorithm == "" && hasher == nil {
			return errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, hashAlgo)
		}

		if !compdesc.IsNoneAccessKind(res.Access.GetKind()) {
			var calculatedDigest *DigestDescriptor
			if (!opts.IsSkipVerify() && digest != "") || (!opts.IsSkipDigest() && digest == "") {
				dig, err := ctx.BlobDigesters().DetermineDigests(res.Type, hasher, opts.HasherProvider, meth, digester)
				if err != nil {
					return err
				}
				if len(dig) == 0 {
					return fmt.Errorf("%s: no digester accepts resource", res.Name)
				}
				calculatedDigest = &dig[0]
			}

			if digest != "" && !opts.IsSkipVerify() {
				if digest != calculatedDigest.Value {
					return fmt.Errorf("digest mismatch: %s != %s", calculatedDigest.Value, digest)
				}
			}

			if !opts.IsSkipDigest() {
				if digest == "" {
					res.Digest = calculatedDigest
				} else {
					res.Digest = &compdesc.DigestSpec{
						HashAlgorithm:          digester.HashAlgorithm,
						NormalisationAlgorithm: digester.NormalizationAlgorithm,
						Value:                  digest,
					}
				}
			}
		}

		if old != nil {
			eq := res.Equivalent(old)
			if !eq.IsLocalHashEqual() && c.impl.IsPersistent() {
				if !opts.IsModifyResource() {
					return fmt.Errorf("resource would invalidate signature")
				}
				cd.Signatures = nil
			}
		}

		if old == nil {
			cd.Resources = append(cd.Resources, *res)
		} else {
			cd.Resources[idx] = *res
		}
		return c.update(false)
	})
}

// evaluateResourceDigest evaluate given potentially partly set digest to determine defaults.
func (c *componentVersionAccessView) evaluateResourceDigest(res, old *compdesc.Resource, opts ModificationOptions) (string, DigesterType, string) {
	var digester DigesterType

	hashAlgo := opts.DefaultHashAlgorithm
	value := ""
	if !res.Digest.IsNone() {
		if res.Digest.IsComplete() {
			value = res.Digest.Value
		}
		if res.Digest.HashAlgorithm != "" {
			hashAlgo = res.Digest.HashAlgorithm
		}
		if res.Digest.NormalisationAlgorithm != "" {
			digester = DigesterType{
				HashAlgorithm:          hashAlgo,
				NormalizationAlgorithm: res.Digest.NormalisationAlgorithm,
			}
		}
	}
	res.Digest = nil

	// keep potential old digest settings
	if old != nil && old.Type == res.Type {
		if !old.Digest.IsNone() {
			digester.HashAlgorithm = old.Digest.HashAlgorithm
			digester.NormalizationAlgorithm = old.Digest.NormalisationAlgorithm
			if opts.IsAcceptExistentDigests() && !opts.IsModifyResource() && c.impl.IsPersistent() {
				res.Digest = old.Digest
				value = old.Digest.Value
			}
		}
	}
	return hashAlgo, digester, value
}

func (c *componentVersionAccessView) SetSourceByAccess(art SourceAccess) error {
	return setAccess(c, "source", art,
		c.SetSource, c.SetSourceBlob)
}

func (c *componentVersionAccessView) SetSource(meta *SourceMeta, acc compdesc.AccessSpec) error {
	if c.impl.IsReadOnly() {
		return accessio.ErrReadOnly
	}

	res := &compdesc.Source{
		SourceMeta: *meta.Copy(),
		Access:     acc,
	}
	return c.Execute(func() error {
		if res.Version == "" {
			res.Version = c.impl.GetVersion()
		}
		cd := c.impl.GetDescriptor()
		if idx := cd.GetSourceIndex(&res.SourceMeta); idx == -1 {
			cd.Sources = append(cd.Sources, *res)
		} else {
			cd.Sources[idx] = *res
		}
		return c.update(false)
	})
}

func (c *componentVersionAccessView) SetReference(ref *ComponentReference) error {
	return c.Execute(func() error {
		cd := c.impl.GetDescriptor()
		if idx := cd.GetComponentReferenceIndex(*ref); idx == -1 {
			cd.References = append(cd.References, *ref)
		} else {
			cd.References[idx] = *ref
		}
		return c.update(false)
	})
}

func (c *componentVersionAccessView) DiscardChanges() {
	c.impl.DiscardChanges()
}

func (c *componentVersionAccessView) IsPersistent() bool {
	return c.impl.IsPersistent()
}

func (c *componentVersionAccessView) UseDirectAccess() bool {
	return c.impl.UseDirectAccess()
}

////////////////////////////////////////////////////////////////////////////////
// Standard Implementation for descriptor based methods

func (c *componentVersionAccessView) GetResource(id metav1.Identity) (ResourceAccess, error) {
	r, err := c.GetDescriptor().GetResourceByIdentity(id)
	if err != nil {
		return nil, err
	}
	return NewResourceAccess(c, r.Access, r.ResourceMeta), nil
}

func (c *componentVersionAccessView) GetResourceIndex(id metav1.Identity) int {
	return c.GetDescriptor().GetResourceIndexByIdentity(id)
}

func (c *componentVersionAccessView) GetResourceByIndex(i int) (ResourceAccess, error) {
	if i < 0 || i >= len(c.GetDescriptor().Resources) {
		return nil, errors.ErrInvalid("resource index", strconv.Itoa(i))
	}
	r := c.GetDescriptor().Resources[i]
	return NewResourceAccess(c, r.Access, r.ResourceMeta), nil
}

func (c *componentVersionAccessView) GetResourcesByName(name string, selectors ...compdesc.IdentitySelector) ([]ResourceAccess, error) {
	resources, err := c.GetDescriptor().GetResourcesByName(name, selectors...)
	if err != nil {
		return nil, err
	}

	result := []ResourceAccess{}
	for _, resource := range resources {
		result = append(result, NewResourceAccess(c, resource.Access, resource.ResourceMeta))
	}
	return result, nil
}

func (c *componentVersionAccessView) GetResources() []ResourceAccess {
	result := []ResourceAccess{}
	for _, r := range c.GetDescriptor().Resources {
		result = append(result, NewResourceAccess(c, r.Access, r.ResourceMeta))
	}
	return result
}

// GetResourcesByIdentitySelectors returns resources that match the given identity selectors.
func (c *componentVersionAccessView) GetResourcesByIdentitySelectors(selectors ...compdesc.IdentitySelector) ([]ResourceAccess, error) {
	return c.GetResourcesBySelectors(selectors, nil)
}

// GetResourcesByResourceSelectors returns resources that match the given resource selectors.
func (c *componentVersionAccessView) GetResourcesByResourceSelectors(selectors ...compdesc.ResourceSelector) ([]ResourceAccess, error) {
	return c.GetResourcesBySelectors(nil, selectors)
}

// GetResourcesBySelectors returns resources that match the given selector.
func (c *componentVersionAccessView) GetResourcesBySelectors(selectors []compdesc.IdentitySelector, resourceSelectors []compdesc.ResourceSelector) ([]ResourceAccess, error) {
	resources := make([]ResourceAccess, 0)
	rscs := c.GetDescriptor().Resources
	for i := range rscs {
		selctx := compdesc.NewResourceSelectionContext(i, rscs)
		if len(selectors) > 0 {
			ok, err := selector.MatchSelectors(selctx.Identity(), selectors...)
			if err != nil {
				return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
			}
			if !ok {
				continue
			}
		}
		ok, err := compdesc.MatchResourceByResourceSelector(selctx, resourceSelectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
		}
		if !ok {
			continue
		}
		r, err := c.GetResourceByIndex(i)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	if len(resources) == 0 {
		return resources, compdesc.NotFound
	}
	return resources, nil
}

func (c *componentVersionAccessView) GetSource(id metav1.Identity) (SourceAccess, error) {
	r, err := c.GetDescriptor().GetSourceByIdentity(id)
	if err != nil {
		return nil, err
	}
	return NewSourceAccess(c, r.Access, r.SourceMeta), nil
}

func (c *componentVersionAccessView) GetSourceIndex(id metav1.Identity) int {
	return c.GetDescriptor().GetSourceIndexByIdentity(id)
}

func (c *componentVersionAccessView) GetSourceByIndex(i int) (SourceAccess, error) {
	if i < 0 || i >= len(c.GetDescriptor().Sources) {
		return nil, errors.ErrInvalid("source index", strconv.Itoa(i))
	}
	r := c.GetDescriptor().Sources[i]
	return NewSourceAccess(c, r.Access, r.SourceMeta), nil
}

func (c *componentVersionAccessView) GetSources() []SourceAccess {
	result := []SourceAccess{}
	for _, r := range c.GetDescriptor().Sources {
		result = append(result, NewSourceAccess(c, r.Access, r.SourceMeta))
	}
	return result
}

func (c *componentVersionAccessView) GetReferences() compdesc.References {
	return c.GetDescriptor().References
}

func (c *componentVersionAccessView) GetReference(id metav1.Identity) (ComponentReference, error) {
	return c.GetDescriptor().GetReferenceByIdentity(id)
}

func (c *componentVersionAccessView) GetReferenceIndex(id metav1.Identity) int {
	return c.GetDescriptor().GetReferenceIndexByIdentity(id)
}

func (c *componentVersionAccessView) GetReferenceByIndex(i int) (ComponentReference, error) {
	if i < 0 || i > len(c.GetDescriptor().References) {
		return ComponentReference{}, errors.ErrInvalid("reference index", strconv.Itoa(i))
	}
	return c.GetDescriptor().References[i], nil
}

func (c *componentVersionAccessView) GetReferencesByName(name string, selectors ...compdesc.IdentitySelector) (compdesc.References, error) {
	return c.GetDescriptor().GetReferencesByName(name, selectors...)
}

// GetReferencesByIdentitySelectors returns references that match the given identity selectors.
func (c *componentVersionAccessView) GetReferencesByIdentitySelectors(selectors ...compdesc.IdentitySelector) (compdesc.References, error) {
	return c.GetReferencesBySelectors(selectors, nil)
}

// GetReferencesByReferenceSelectors returns references that match the given resource selectors.
func (c *componentVersionAccessView) GetReferencesByReferenceSelectors(selectors ...compdesc.ReferenceSelector) (compdesc.References, error) {
	return c.GetReferencesBySelectors(nil, selectors)
}

// GetReferencesBySelectors returns references that match the given selector.
func (c *componentVersionAccessView) GetReferencesBySelectors(selectors []compdesc.IdentitySelector, referenceSelectors []compdesc.ReferenceSelector) (compdesc.References, error) {
	references := make(compdesc.References, 0)
	refs := c.GetDescriptor().References
	for i := range refs {
		selctx := compdesc.NewReferenceSelectionContext(i, refs)
		if len(selectors) > 0 {
			ok, err := selector.MatchSelectors(selctx.Identity(), selectors...)
			if err != nil {
				return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
			}
			if !ok {
				continue
			}
		}
		ok, err := compdesc.MatchReferencesByReferenceSelector(selctx, referenceSelectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
		}
		if !ok {
			continue
		}
		references = append(references, *selctx.ComponentReference)
	}
	if len(references) == 0 {
		return references, compdesc.NotFound
	}
	return references, nil
}
