package repocpi

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/compose"
	"ocm.software/ocm/api/ocm/extensions/attrs/compositionmodeattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/keepblobattr"
	"ocm.software/ocm/api/ocm/extensions/pubsub"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/refmgmt/resource"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

// here, we define the common implementation agnostic parts
// for component version objects referred to by a ComponentVersionView.

// ComponentVersionAccessImpl is the provider implementation
// interface for component versions.
type ComponentVersionAccessImpl interface {
	GetContext() cpi.Context
	SetBridge(bridge ComponentVersionAccessBridge)
	GetParentBridge() ComponentAccessBridge

	Repository() cpi.Repository

	IsReadOnly() bool
	SetReadOnly()

	GetDescriptor() *compdesc.ComponentDescriptor
	SetDescriptor(*compdesc.ComponentDescriptor) (bool, error)

	AccessMethod(acc cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error)

	BlobContainer
	io.Closer
}

type _componentVersionAccessBridgeBase = resource.ResourceImplBase[cpi.ComponentVersionAccess]

// componentVersionAccessBridge is the counterpart to views, all views
// created by Dup calls use this base object to work on.
// Besides some functionality covered by view objects these base objects
// implement provider-agnostic parts of the ComponentVersionAccess API.
type componentVersionAccessBridge struct {
	lock sync.Mutex
	id   runtimefinalizer.ObjectIdentity

	*_componentVersionAccessBridgeBase
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

var _ ComponentVersionAccessBridge = (*componentVersionAccessBridge)(nil)

func newComponentVersionAccessBridge(name, version string, impl ComponentVersionAccessImpl, lazy, persistent, direct bool, closer ...io.Closer) (ComponentVersionAccessBridge, error) {
	base, err := resource.NewResourceImplBase[cpi.ComponentVersionAccess, cpi.ComponentAccess](impl.GetParentBridge(), closer...)
	if err != nil {
		return nil, err
	}
	b := &componentVersionAccessBridge{
		_componentVersionAccessBridgeBase: base,
		id:                                runtimefinalizer.NewObjectIdentity(fmt.Sprintf("%s:%s", name, version)),
		ctx:                               impl.GetContext(),
		name:                              name,
		version:                           version,
		blobcache:                         NewBlobCache(),
		lazy:                              lazy,
		persistent:                        persistent,
		directAccess:                      direct,
		impl:                              impl,
	}
	impl.SetBridge(b)
	return b, nil
}

func GetComponentVersionImpl[T ComponentVersionAccessImpl](cv cpi.ComponentVersionAccess) (T, error) {
	var _nil T

	impl, err := GetComponentVersionAccessBridge(cv)
	if err != nil {
		return _nil, err
	}
	if mine, ok := impl.(*componentVersionAccessBridge); ok {
		cont, ok := mine.impl.(T)
		if ok {
			return cont, nil
		}
		return _nil, errors.Newf("non-matching component version implementation %T", mine.impl)
	}
	return _nil, errors.Newf("non-matching component version implementation %T", impl)
}

func (b *componentVersionAccessBridge) Close() error {
	list := errors.ErrListf("closing component version %s", common.VersionedElementKey(b))
	refmgmt.AllocLog.Trace("closing component version base", "name", common.VersionedElementKey(b))
	// prepare artifact access for final close in
	// direct access mode.
	if !compositionmodeattr.Get(b.ctx) {
		list.Add(b.update(true))
	}
	list.Add(b.impl.Close())
	list.Add(b._componentVersionAccessBridgeBase.Close())
	list.Add(b.blobcache.Clear())
	refmgmt.AllocLog.Trace("closed component version base", "name", common.VersionedElementKey(b))
	return list.Result()
}

func (b *componentVersionAccessBridge) GetContext() cpi.Context {
	return b.ctx
}

func (b *componentVersionAccessBridge) GetName() string {
	return b.name
}

func (b *componentVersionAccessBridge) GetVersion() string {
	return b.version
}

func (b *componentVersionAccessBridge) GetImplementation() ComponentVersionAccessImpl {
	return b.impl
}

func (b *componentVersionAccessBridge) GetBlobCache() BlobCache {
	return b.blobcache
}

func (b *componentVersionAccessBridge) EnablePersistence() bool {
	if b.discardChanges {
		return false
	}
	b.persistent = true
	b.GetStorageContext()
	return true
}

func (b *componentVersionAccessBridge) IsPersistent() bool {
	return b.persistent
}

func (b *componentVersionAccessBridge) UseDirectAccess() bool {
	return b.directAccess
}

func (b *componentVersionAccessBridge) DiscardChanges() {
	b.discardChanges = true
}

func (b *componentVersionAccessBridge) Repository() cpi.Repository {
	return b.impl.Repository()
}

func (b *componentVersionAccessBridge) IsReadOnly() bool {
	return b.impl.IsReadOnly()
}

func (b *componentVersionAccessBridge) SetReadOnly() {
	b.impl.SetReadOnly()
}

////////////////////////////////////////////////////////////////////////////////
// with access to actual view

func (b *componentVersionAccessBridge) AccessMethod(spec cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (meth cpi.AccessMethod, err error) {
	switch {
	case compose.Is(spec):
		cspec, ok := spec.(*compose.AccessSpec)
		if !ok {
			return nil, fmt.Errorf("invalid implementation (%T) for access method compose", spec)
		}
		blob := b.getLocalBlob(cspec)
		if blob == nil {
			return nil, errors.ErrUnknown(blobaccess.KIND_BLOB, cspec.Id, common.VersionedElementKey(b).String())
		}
		meth, err = compose.NewMethod(cspec, blob)
	case spec.IsLocal(b.ctx):
		meth, err = b.impl.AccessMethod(spec, cv)
		if err == nil {
			if blob := b.getLocalBlob(spec); blob != nil {
				meth, err = newFakeMethod(meth, blob)
			}
		}
	}
	return meth, err
}

func (b *componentVersionAccessBridge) GetDescriptor() *compdesc.ComponentDescriptor {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.getDescriptor()
}

func (b *componentVersionAccessBridge) getDescriptor() *compdesc.ComponentDescriptor {
	if b.descriptor == nil {
		b.descriptor = b.impl.GetDescriptor()
	}
	return b.descriptor
}

func (b *componentVersionAccessBridge) GetStorageContext() cpi.StorageContext {
	return b.impl.GetStorageContext()
}

func (b *componentVersionAccessBridge) ShouldUpdate(final bool) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.shouldUpdate(final)
}

func (b *componentVersionAccessBridge) shouldUpdate(final bool) bool {
	if b.discardChanges {
		return false
	}
	if final {
		return b.persistent
	}
	return !b.lazy && b.directAccess && b.persistent
}

func (b *componentVersionAccessBridge) Update(final bool) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.update(final)
}

func (b *componentVersionAccessBridge) update(final bool) error {
	if !b.shouldUpdate(final) {
		return nil
	}

	d := b.getDescriptor()

	opts := &cpi.BlobUploadOptions{
		UseNoDefaultIfNotSet: optionutils.PointerTo(true),
	}
	err := b.setupLocalBlobs("resource", b.composeAccess, d.Resources, true, opts)
	if err == nil {
		err = b.setupLocalBlobs("source", b.composeAccess, d.Sources, true, opts)
	}
	if err != nil {
		return err
	}

	updated, err := b.impl.SetDescriptor(b.descriptor.Copy())
	if err != nil {
		return err
	}
	err = b.blobcache.Clear()

	if updated {
		pub, err := pubsub.PubSubForRepo(b.Repository())
		if err != nil {
			return err
		}
		if pub != nil {
			err := pub.NotifyComponentVersion(common.VersionedElementKey(b))
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (b *componentVersionAccessBridge) getLocalBlob(acc cpi.AccessSpec) cpi.BlobAccess {
	key, err := json.Marshal(acc)
	if err != nil {
		return nil
	}
	return b.blobcache.GetBlobFor(string(key))
}

func (b *componentVersionAccessBridge) AddBlob(blob cpi.BlobAccess, artType, refName string, global cpi.AccessSpec, final bool, opts *cpi.BlobUploadOptions) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	if b.IsReadOnly() {
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

	ctx := b.GetContext()

	// handle foreign blob upload
	var prov cpi.BlobHandlerProvider
	if opts.BlobHandlerProvider != nil {
		prov = opts.BlobHandlerProvider
	} else {
		if !optionutils.AsValue(opts.UseNoDefaultIfNotSet) {
			prov = internal.BlobHandlerProviderForRegistry(ctx.BlobHandlers())
		}
	}
	if prov != nil {
		storagectx := b.GetStorageContext()
		mime := blob.MimeType()
		h := prov.LookupHandler(storagectx, artType, mime)
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

	if final || b.UseDirectAccess() {
		acc, err = b.impl.AddBlob(blob, refName, global)
		if err != nil {
			return nil, err
		}
	} else {
		// use local composition access to be added to the repository with AddVersion.
		acc = compose.New(refName, blob.MimeType(), global)
	}
	return b.cacheLocalBlob(acc, blob)
}

func (b *componentVersionAccessBridge) cacheLocalBlob(acc cpi.AccessSpec, blob cpi.BlobAccess) (cpi.AccessSpec, error) {
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
	k := BlobCacheKey(string(key))
	err = b.blobcache.AddBlobFor(k, blob)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

////////////////////////////////////////////////////////////////////////////////

func (b *componentVersionAccessBridge) composeAccess(spec cpi.AccessSpec) (blobaccess.BlobAccess, string, cpi.AccessSpec, error) {
	if !compose.Is(spec) {
		return nil, "", nil, nil
	}
	cspec, ok := spec.(*compose.AccessSpec)
	if !ok {
		return nil, "", nil, fmt.Errorf("invalid implementation (%T) for access method compose", spec)
	}
	blob := b.getLocalBlob(cspec)
	if blob == nil {
		return nil, "", nil, errors.ErrUnknown(blobaccess.KIND_BLOB, cspec.Id, common.VersionedElementKey(b).String())
	}
	blob, err := blob.Dup()
	if err != nil {
		return nil, "", nil, errors.Wrapf(err, "cached blob")
	}

	return blob, cspec.ReferenceName, cspec.GlobalAccess.Get(), nil
}

func (b *componentVersionAccessBridge) setupLocalBlobs(kind string, accprov func(cpi.AccessSpec) (blobaccess.BlobAccess, string, cpi.AccessSpec, error), it compdesc.ArtifactAccessor, final bool, opts *cpi.BlobUploadOptions) (ferr error) {
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&ferr)

	for i := 0; i < it.Len(); i++ {
		nested := finalize.Nested()
		a := it.GetArtifact(i)
		spec, err := b.ctx.AccessSpecForSpec(a.GetAccess())
		if err != nil {
			return errors.Wrapf(err, "%s %d", kind, i)
		}
		blob, ref, global, err := accprov(spec)
		if err != nil {
			return errors.Wrapf(err, "%s %d", kind, i)
		}
		if blob != nil {
			nested.Close(blob)

			effspec, err := b.AddBlob(blob, a.GetType(), ref, global, final, opts)
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

////////////////////////////////////////////////////////////////////////////////

type fakeMethod struct {
	spec  cpi.AccessSpec
	local bool
	mime  string
	blob  blobaccess.BlobAccess
}

var _ accspeccpi.AccessMethodImpl = (*fakeMethod)(nil)

func newFakeMethod(m cpi.AccessMethod, blob cpi.BlobAccess) (cpi.AccessMethod, error) {
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
	return accspeccpi.AccessMethodForImplementation(f, nil)
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
