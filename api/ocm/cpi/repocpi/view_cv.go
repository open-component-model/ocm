package repocpi

import (
	"fmt"
	"io"
	"strconv"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/ocm/selectors/refsel"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/ocm/selectors/srcsel"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/refmgmt/resource"
)

// View objects are the user facing generic implementations of the context interfaces.
// They are responsible to handle the reference counting and use
// shared implementations objects for th concrete type-specific implementations.
// Additionally, they are used to implement interface functionality which is
// common to all implementations and NOT dependent on the backend system technology.

// here are the views implementing the user facing ComponentVersionAccess
// interface.

type _componentVersionAccessView interface {
	resource.ResourceViewInt[cpi.ComponentVersionAccess]
}

type ComponentVersionAccessViewManager = resource.ViewManager[cpi.ComponentVersionAccess]

type ComponentVersionAccessBridge interface {
	resource.ResourceImplementation[cpi.ComponentVersionAccess]
	common.VersionedElement
	io.Closer

	GetContext() cpi.Context
	Repository() cpi.Repository

	GetImplementation() ComponentVersionAccessImpl

	EnablePersistence() bool
	DiscardChanges()
	IsPersistent() bool

	GetDescriptor() *compdesc.ComponentDescriptor

	AccessMethod(cpi.AccessSpec, refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error)

	// GetStorageContext creates a storage context for blobs
	// that is used to feed blob handlers for specific blob storage methods.
	// If no handler accepts the blob, the AddBlob method will
	// be used to store the blob
	GetStorageContext() cpi.StorageContext

	// AddBlob stores a local blob together with the component and
	// potentially provides a global reference.
	// The resulting access information (global and local) is provided as
	// an access method specification usable in a component descriptor.
	// This is the direct technical storage, without caring about any handler.
	AddBlob(blob cpi.BlobAccess, arttype, refName string, global cpi.AccessSpec, final bool, opts *cpi.BlobUploadOptions) (cpi.AccessSpec, error)

	IsReadOnly() bool
	SetReadOnly()

	// ShouldUpdate checks, whether an update is indicated
	// by the state of object, considering persistence, lazy, discard
	// and update mode state
	ShouldUpdate(final bool) bool

	// GetBlobCache retrieves the blob cache used to store preliminary
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

type componentVersionAccessView struct {
	_componentVersionAccessView
	bridge ComponentVersionAccessBridge
	err    error
}

var (
	_ cpi.ComponentVersionAccess = (*componentVersionAccessView)(nil)
	_ utils.Unwrappable          = (*componentVersionAccessView)(nil)
)

func GetComponentVersionAccessBridge(n cpi.ComponentVersionAccess) (ComponentVersionAccessBridge, error) {
	if v, ok := n.(*componentVersionAccessView); ok {
		return v.bridge, nil
	}
	return nil, errors.ErrNotSupported("component version bridge type", fmt.Sprintf("%T", n))
}

func GetComponentVersionAccessImplementation(n cpi.ComponentVersionAccess) (ComponentVersionAccessImpl, error) {
	if v, ok := n.(*componentVersionAccessView); ok {
		if b, ok := v.bridge.(*componentVersionAccessBridge); ok {
			return b.impl, nil
		}
		return nil, errors.ErrNotSupported("component version bridge type", fmt.Sprintf("%T", v.bridge))
	}
	return nil, errors.ErrNotSupported("component version implementation type", fmt.Sprintf("%T", n))
}

func artifactAccessViewCreator(i ComponentVersionAccessBridge, v resource.CloserView, d resource.ViewManager[cpi.ComponentVersionAccess]) cpi.ComponentVersionAccess {
	cv := &componentVersionAccessView{
		_componentVersionAccessView: resource.NewView[cpi.ComponentVersionAccess](v, d),
		bridge:                      i,
	}
	return cv
}

func NewComponentVersionAccess(name, version string, impl ComponentVersionAccessImpl, lazy, persistent, direct bool, closer ...io.Closer) (cpi.ComponentVersionAccess, error) {
	bridge, err := newComponentVersionAccessBridge(name, version, impl, lazy, persistent, direct, closer...)
	if err != nil {
		return nil, errors.Join(err, impl.Close())
	}
	return resource.NewResource[cpi.ComponentVersionAccess](bridge, artifactAccessViewCreator, fmt.Sprintf("component version  %s/%s", name, version), true), nil
}

func (c *componentVersionAccessView) Unwrap() interface{} {
	return c.bridge
}

func (c *componentVersionAccessView) IsReadOnly() bool {
	return c.bridge.IsReadOnly()
}

func (c *componentVersionAccessView) SetReadOnly() {
	c.bridge.SetReadOnly()
}

func (c *componentVersionAccessView) Close() error {
	list := errors.ErrListf("closing %s", common.VersionedElementKey(c))
	err := c._componentVersionAccessView.Close()
	return list.Add(c.err, err).Result()
}

func (c *componentVersionAccessView) Repository() cpi.Repository {
	return c.bridge.Repository()
}

func (c *componentVersionAccessView) GetContext() cpi.Context {
	return c.bridge.GetContext()
}

func (c *componentVersionAccessView) GetName() string {
	return c.bridge.GetName()
}

func (c *componentVersionAccessView) GetVersion() string {
	return c.bridge.GetVersion()
}

func (c *componentVersionAccessView) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.bridge.GetDescriptor()
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

func (c *componentVersionAccessView) AccessMethod(spec cpi.AccessSpec) (meth cpi.AccessMethod, err error) {
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

func (c *componentVersionAccessView) accessMethod(spec cpi.AccessSpec) (meth cpi.AccessMethod, err error) {
	switch {
	case spec.IsLocal(c.GetContext()):
		return c.bridge.AccessMethod(spec, c.Allocatable())
	default:
		return spec.AccessMethod(c)
	}
}

func (c *componentVersionAccessView) Update() error {
	return c.Execute(func() error {
		if !c.bridge.IsPersistent() {
			return ErrTempVersion
		}
		return c.bridge.Update(true)
	})
}

func (c *componentVersionAccessView) AddBlob(blob cpi.BlobAccess, artType, refName string, global cpi.AccessSpec, opts ...cpi.BlobUploadOption) (cpi.AccessSpec, error) {
	var spec cpi.AccessSpec
	eff := cpi.NewBlobUploadOptions(opts...)
	err := c.Execute(func() error {
		var err error
		spec, err = c.bridge.AddBlob(blob, artType, refName, global, false, eff)
		return err
	})

	return spec, err
}

func (c *componentVersionAccessView) AdjustResourceAccess(meta *cpi.ResourceMeta, acc compdesc.AccessSpec, opts ...cpi.ModificationOption) error {
	cd := c.GetDescriptor()
	if idx := cd.GetResourceIndex(meta); idx >= 0 {
		return c.SetResource(&cd.Resources[idx].ResourceMeta, acc, opts...)
	}
	return errors.ErrUnknown(cpi.KIND_RESOURCE, meta.GetIdentity(cd.Resources).String())
}

// SetResourceBlob adds a blob resource to the component version.
func (c *componentVersionAccessView) SetResourceBlob(meta *cpi.ResourceMeta, blob cpi.BlobAccess, refName string, global cpi.AccessSpec, opts ...cpi.BlobModificationOption) error {
	cpi.Logger(c).Debug("adding resource blob", "resource", meta.Name)
	if err := utils.ValidateObject(blob); err != nil {
		return err
	}
	eff := cpi.NewBlobModificationOptions(opts...)
	acc, err := c.AddBlob(blob, meta.Type, refName, global, eff)
	if err != nil {
		return fmt.Errorf("unable to add blob (component %s:%s resource %s): %w", c.GetName(), c.GetVersion(), meta.GetName(), err)
	}

	if err := c.SetResource(meta, acc, eff, cpi.ModifyElement()); err != nil {
		return fmt.Errorf("unable to set resource: %w", err)
	}

	return nil
}

func (c *componentVersionAccessView) AdjustSourceAccess(meta *cpi.SourceMeta, acc compdesc.AccessSpec) error {
	cd := c.GetDescriptor()
	if idx := cd.GetSourceIndex(meta); idx >= 0 {
		return c.SetSource(&cd.Sources[idx].SourceMeta, acc)
	}
	return errors.ErrUnknown(cpi.KIND_RESOURCE, meta.GetIdentity(cd.Resources).String())
}

func (c *componentVersionAccessView) SetSourceBlob(meta *cpi.SourceMeta, blob cpi.BlobAccess, refName string, global cpi.AccessSpec, modopts ...cpi.TargetElementOption) error {
	cpi.Logger(c).Debug("adding source blob", "source", meta.Name)
	if err := utils.ValidateObject(blob); err != nil {
		return err
	}
	acc, err := c.AddBlob(blob, meta.Type, refName, global)
	if err != nil {
		return fmt.Errorf("unable to add blob: (component %s:%s source %s): %w", c.GetName(), c.GetVersion(), meta.GetName(), err)
	}

	if err := c.SetSource(meta, acc, modopts...); err != nil {
		return fmt.Errorf("unable to set source: %w", err)
	}

	return nil
}

func setAccess[T any, A cpi.ArtifactAccess[T]](c *componentVersionAccessView, kind string, art A,
	set func(*T, compdesc.AccessSpec) error,
	setblob func(*T, cpi.BlobAccess, string, cpi.AccessSpec) error,
) error {
	if c.bridge.IsReadOnly() {
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
		blob   cpi.BlobAccess
		hint   string
		global cpi.AccessSpec
	)

	if acc != nil {
		if !acc.IsLocal(c.GetContext()) {
			return set(meta, acc)
		}

		blob, err = accspeccpi.BlobAccessForAccessSpec(acc, c)
		if err != nil && errors.IsErrNotFoundElem(err, "", blobaccess.KIND_BLOB) {
			return err
		}
		hint = cpi.ReferenceHint(acc, c)
		global = cpi.GlobalAccess(acc, c.GetContext())
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

func (c *componentVersionAccessView) SetResourceByAccess(art cpi.ResourceAccess, modopts ...cpi.BlobModificationOption) error {
	return setAccess(c, "resource", art,
		func(meta *cpi.ResourceMeta, acc compdesc.AccessSpec) error {
			return c.SetResource(meta, acc, cpi.NewBlobModificationOptions(modopts...))
		},
		func(meta *cpi.ResourceMeta, blob cpi.BlobAccess, hint string, global cpi.AccessSpec) error {
			return c.SetResourceBlob(meta, blob, hint, global, modopts...)
		})
}

func (c *componentVersionAccessView) SetResource(meta *cpi.ResourceMeta, acc compdesc.AccessSpec, modopts ...cpi.ModificationOption) error {
	if c.bridge.IsReadOnly() {
		return accessio.ErrReadOnly
	}

	res := &compdesc.Resource{
		ResourceMeta: *meta.Copy(),
		Access:       acc,
	}

	ctx := c.bridge.GetContext()
	opts := cpi.NewModificationOptions(modopts...)
	cpi.CompleteModificationOptions(ctx, opts)

	spec, err := c.bridge.GetContext().AccessSpecForSpec(acc)
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
	defer meth.Close()

	return c.Execute(func() error {
		var old *compdesc.Resource

		if res.Relation == metav1.LocalRelation {
			if res.Version == "" {
				res.Version = c.GetVersion()
			}
		}

		cd := c.bridge.GetDescriptor()

		idx, err := c.getElementIndex("resource", cd.Resources, res, &opts.TargetElementOptions)
		if err != nil {
			return err
		}
		if idx >= 0 {
			old = &cd.Resources[idx]
		}

		if old == nil {
			if !opts.IsModifyElement() && c.bridge.IsPersistent() {
				return fmt.Errorf("new resource would invalidate signature")
			}
		}

		// evaluate given digesting constraints and settings
		hashAlgo, digester, digest := c.evaluateResourceDigest(res, old, *opts)
		digestForwarded := false
		if digest == "" {
			if p, ok := meth.(DigestSpecProvider); ok {
				dig, err := p.GetDigestSpec()
				if dig != nil && err == nil {
					// always prefer already known digest with its method
					// if no concrete digest value is given by the caller
					digest = dig.Value
					hashAlgo = dig.HashAlgorithm
					digester.HashAlgorithm = hashAlgo
					digester.NormalizationAlgorithm = dig.NormalisationAlgorithm
					digestForwarded = true
				}
			}
		}

		hasher := opts.GetHasher(hashAlgo)
		if hasher == nil {
			return errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, hashAlgo)
		}

		if !compdesc.IsNoneAccessKind(res.Access.GetKind()) {
			var calculatedDigest *cpi.DigestDescriptor
			if (!opts.IsSkipVerify() && !digestForwarded && digest != "") || (!opts.IsSkipDigest() && digest == "") {
				dig, err := ctx.BlobDigesters().DetermineDigests(res.Type, hasher, opts.HasherProvider, meth, digester)
				if err != nil {
					return err
				}
				if len(dig) == 0 {
					return fmt.Errorf("%s: no digester accepts resource", res.Name)
				}
				calculatedDigest = &dig[0]

				if digest != "" && !opts.IsSkipVerify() {
					if digest != calculatedDigest.Value {
						return fmt.Errorf("digest mismatch: %s != %s", calculatedDigest.Value, digest)
					}
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
			if !eq.IsLocalHashEqual() && c.bridge.IsPersistent() {
				if !opts.IsModifyElement() {
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
		if opts.IsModifyElement() && !opts.IsDisableExtraIdentityDefaulting() {
			// default handling for completing an extra identity for modifications, only.
			compdesc.DefaultResources(cd)
		}
		return c.bridge.Update(false)
	})
}

// evaluateResourceDigest evaluate given potentially partly set digest to determine defaults.
func (c *componentVersionAccessView) evaluateResourceDigest(res, old *compdesc.Resource, opts cpi.ModificationOptions) (string, cpi.DigesterType, string) {
	var digester cpi.DigesterType

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
			digester = cpi.DigesterType{
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
			if opts.IsAcceptExistentDigests() && !opts.IsModifyElement() && c.bridge.IsPersistent() {
				res.Digest = old.Digest
				value = old.Digest.Value
			}
		}
	}
	return hashAlgo, digester, value
}

func (c *componentVersionAccessView) SetSourceByAccess(art cpi.SourceAccess, optslist ...cpi.TargetElementOption) error {
	return setAccess(c, "source", art,
		func(meta *cpi.SourceMeta, acc compdesc.AccessSpec) error {
			return c.SetSource(meta, acc, optslist...)
		},
		func(meta *cpi.SourceMeta, blob cpi.BlobAccess, hint string, global cpi.AccessSpec) error {
			return c.SetSourceBlob(meta, blob, hint, global, optslist...)
		})
}

func (c *componentVersionAccessView) SetSource(meta *cpi.SourceMeta, acc compdesc.AccessSpec, optlist ...cpi.TargetElementOption) error {
	if c.bridge.IsReadOnly() {
		return accessio.ErrReadOnly
	}

	opts := cpi.NewTargetElementOptions(optlist...)
	res := &compdesc.Source{
		SourceMeta: *meta.Copy(),
		Access:     acc,
	}

	return c.Execute(func() error {
		if res.Version == "" {
			res.Version = c.bridge.GetVersion()
		}
		cd := c.bridge.GetDescriptor()

		idx, err := c.getElementIndex("source", cd.Sources, res, optlist...)
		if err != nil {
			return err
		}

		if idx < 0 {
			cd.Sources = append(cd.Sources, *res)
		} else {
			cd.Sources[idx] = *res
		}
		if !opts.IsDisableExtraIdentityDefaulting() {
			compdesc.DefaultSources(cd)
		}
		return c.bridge.Update(false)
	})
}

func (c *componentVersionAccessView) SetReference(ref *cpi.ComponentReference, optlist ...cpi.ElementModificationOption) error {
	opts := cpi.NewElementModificationOptions(optlist...)
	moddef := false

	return c.Execute(func() error {
		cd := c.bridge.GetDescriptor()

		if ref.Version == "" {
			return fmt.Errorf("version required for component version reference")
		}
		idx, err := c.getElementIndex("reference", cd.References, ref, &opts.TargetElementOptions)
		if err != nil {
			return err
		}

		if idx < 0 {
			if !opts.IsModifyElement(moddef) {
				return fmt.Errorf("adding reference would invalidate signature")
			}
			cd.References = append(cd.References, *ref)
		} else {
			eq := ref.Equivalent(&cd.References[idx])
			if !eq.IsEquivalent() && c.bridge.IsPersistent() {
				if !opts.IsModifyElement(moddef) {
					return fmt.Errorf("reference would invalidate signature")
				}
				cd.Signatures = nil
			}
			cd.References[idx].Equivalent(ref)
			cd.References[idx] = *ref
		}
		if opts.IsModifyElement(moddef) && !opts.IsDisableExtraIdentityDefaulting() {
			compdesc.DefaultReferences(cd)
		}
		return c.bridge.Update(false)
	})
}

func (c *componentVersionAccessView) getElementIndex(kind string, acc compdesc.ElementListAccessor, prov compdesc.ElementMetaProvider, optlist ...cpi.TargetElementOption) (int, error) {
	opts := internal.NewTargetElementOptions(optlist...)
	curidx := compdesc.ElementIndex(acc, prov)
	meta := prov.GetMeta()
	var idx int
	if opts.TargetElement != nil {
		var err error
		idx, err = opts.TargetElement.GetTargetIndex(acc, meta)
		if err != nil {
			return idx, err
		}
		if idx == -1 && curidx >= 0 {
			if meta.GetVersion() == acc.Get(curidx).GetMeta().GetVersion() {
				return -1, fmt.Errorf("adding a new %s with same base identity requires different version", kind)
			}
		}
		if idx >= acc.Len() {
			return -1, fmt.Errorf("index %d out of range of %s list", idx, kind)
		}
		if idx < -1 {
			return -1, fmt.Errorf("invalid index %d for %s list", idx, kind)
		}
	} else {
		idx = curidx
	}
	return idx, nil
}

func (c *componentVersionAccessView) DiscardChanges() {
	c.bridge.DiscardChanges()
}

func (c *componentVersionAccessView) IsPersistent() bool {
	return c.bridge.IsPersistent()
}

func (c *componentVersionAccessView) UseDirectAccess() bool {
	return c.bridge.UseDirectAccess()
}

////////////////////////////////////////////////////////////////////////////////
// Standard Implementation for descriptor based methods

func (c *componentVersionAccessView) GetResource(id metav1.Identity) (cpi.ResourceAccess, error) {
	r, err := c.GetDescriptor().GetResourceByIdentity(id)
	if err != nil {
		return nil, err
	}
	return cpi.NewResourceAccess(c, r.Access, r.ResourceMeta), nil
}

func (c *componentVersionAccessView) GetResourceIndex(id metav1.Identity) int {
	return c.GetDescriptor().GetResourceIndexByIdentity(id)
}

func (c *componentVersionAccessView) GetResourceByIndex(i int) (cpi.ResourceAccess, error) {
	if i < 0 || i >= len(c.GetDescriptor().Resources) {
		return nil, errors.ErrInvalid("resource index", strconv.Itoa(i))
	}
	r := c.GetDescriptor().Resources[i]
	return cpi.NewResourceAccess(c, r.Access, r.ResourceMeta), nil
}

func (c *componentVersionAccessView) SelectResources(sel ...rscsel.Selector) ([]cpi.ResourceAccess, error) {
	err := selectors.ValidateSelectors(sel...)
	if err != nil {
		return nil, err
	}

	list := compdesc.MapToSelectorElementList(c.GetDescriptor().Resources)
	result := []cpi.ResourceAccess{}
outer:
	for _, r := range c.GetDescriptor().Resources {
		if len(sel) > 0 {
			mr := compdesc.MapToSelectorResource(&r)
			for _, s := range sel {
				if !s.MatchResource(list, mr) {
					continue outer
				}
			}
		}
		result = append(result, cpi.NewResourceAccess(c, r.Access, r.ResourceMeta))
	}
	return result, nil
}

func (c *componentVersionAccessView) GetResources() []cpi.ResourceAccess {
	result := []cpi.ResourceAccess{}
	for _, r := range c.GetDescriptor().Resources {
		result = append(result, cpi.NewResourceAccess(c, r.Access, r.ResourceMeta))
	}
	return result
}

func (c *componentVersionAccessView) GetSource(id metav1.Identity) (cpi.SourceAccess, error) {
	r, err := c.GetDescriptor().GetSourceByIdentity(id)
	if err != nil {
		return nil, err
	}
	return cpi.NewSourceAccess(c, r.Access, r.SourceMeta), nil
}

func (c *componentVersionAccessView) GetSourceIndex(id metav1.Identity) int {
	return c.GetDescriptor().GetSourceIndexByIdentity(id)
}

func (c *componentVersionAccessView) GetSourceByIndex(i int) (cpi.SourceAccess, error) {
	if i < 0 || i >= len(c.GetDescriptor().Sources) {
		return nil, errors.ErrInvalid("source index", strconv.Itoa(i))
	}
	r := c.GetDescriptor().Sources[i]
	return cpi.NewSourceAccess(c, r.Access, r.SourceMeta), nil
}

func (c *componentVersionAccessView) SelectSources(sel ...srcsel.Selector) ([]cpi.SourceAccess, error) {
	err := selectors.ValidateSelectors(sel...)
	if err != nil {
		return nil, err
	}

	list := compdesc.MapToSelectorElementList(c.GetDescriptor().Sources)
	result := []cpi.SourceAccess{}
outer:
	for _, r := range c.GetDescriptor().Sources {
		if len(sel) > 0 {
			mr := compdesc.MapToSelectorSource(&r)
			for _, s := range sel {
				if !s.MatchSource(list, mr) {
					continue outer
				}
			}
		}
		result = append(result, cpi.NewSourceAccess(c, r.Access, r.SourceMeta))
	}
	return result, nil
}

func (c *componentVersionAccessView) GetSources() []cpi.SourceAccess {
	result := []cpi.SourceAccess{}
	for _, r := range c.GetDescriptor().Sources {
		result = append(result, cpi.NewSourceAccess(c, r.Access, r.SourceMeta))
	}
	return result
}

func (c *componentVersionAccessView) SelectReferences(sel ...refsel.Selector) ([]compdesc.Reference, error) {
	err := selectors.ValidateSelectors(sel...)
	if err != nil {
		return nil, err
	}
	return c.GetDescriptor().SelectReferences(sel...)
}

func (c *componentVersionAccessView) GetReferences() []compdesc.Reference {
	return c.GetDescriptor().GetReferences()
}

func (c *componentVersionAccessView) GetReference(id metav1.Identity) (cpi.ComponentReference, error) {
	return c.GetDescriptor().GetReferenceByIdentity(id)
}

func (c *componentVersionAccessView) GetReferenceIndex(id metav1.Identity) int {
	return c.GetDescriptor().GetReferenceIndexByIdentity(id)
}

func (c *componentVersionAccessView) GetReferenceByIndex(i int) (cpi.ComponentReference, error) {
	if i < 0 || i > len(c.GetDescriptor().References) {
		return cpi.ComponentReference{}, errors.ErrInvalid("reference index", strconv.Itoa(i))
	}
	return c.GetDescriptor().References[i], nil
}
