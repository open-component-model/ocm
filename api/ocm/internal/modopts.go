package internal

import (
	"fmt"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/ocm/compdesc"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors/accessors"
	"ocm.software/ocm/api/utils"
)

type BlobUploadOption interface {
	ApplyBlobUploadOption(opts *BlobUploadOptions)
}

type BlobOptionImpl interface {
	BlobUploadOption
	BlobModificationOption
}

type BlobUploadOptions struct {
	UseNoDefaultIfNotSet *bool               `json:"noDefaultUpload,omitempty"`
	BlobHandlerProvider  BlobHandlerProvider `json:"-"`
}

var _ BlobUploadOption = (*BlobUploadOptions)(nil)

func NewBlobUploadOptions(list ...BlobUploadOption) *BlobUploadOptions {
	var m BlobUploadOptions
	m.ApplyBlobUploadOptions(list...)
	return &m
}

func (m *BlobUploadOptions) ApplyBlobUploadOptions(list ...BlobUploadOption) {
	for _, o := range list {
		if o != nil {
			o.ApplyBlobUploadOption(m)
		}
	}
}

func (o *BlobUploadOptions) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	o.ApplyBlobUploadOption(&opts.BlobUploadOptions)
}

func (o *BlobUploadOptions) ApplyBlobUploadOption(opts *BlobUploadOptions) {
	optionutils.ApplyOption(o.UseNoDefaultIfNotSet, &opts.UseNoDefaultIfNotSet)
	if o.BlobHandlerProvider != nil {
		opts.BlobHandlerProvider = o.BlobHandlerProvider
		opts.UseNoDefaultIfNotSet = utils.BoolP(true)
	}
}

////////////////////////////////////////////////////////////////////////////////

type nodefaulthandler bool

func (o nodefaulthandler) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	o.ApplyBlobUploadOption(&opts.BlobUploadOptions)
}

func (o nodefaulthandler) ApplyBlobUploadOption(opts *BlobUploadOptions) {
	opts.UseNoDefaultIfNotSet = optionutils.PointerTo(bool(o))
}

func UseNoDefaultBlobHandlers(b ...bool) BlobOptionImpl {
	return nodefaulthandler(utils.OptionalDefaultedBool(true, b...))
}

////////////////////////////////////////////////////////////////////////////////

type handler struct {
	blobHandlerProvider BlobHandlerProvider
}

func (o *handler) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	o.ApplyBlobUploadOption(&opts.BlobUploadOptions)
}

func (o *handler) ApplyBlobUploadOption(opts *BlobUploadOptions) {
	if o.blobHandlerProvider != nil {
		opts.BlobHandlerProvider = o.blobHandlerProvider
	}
}

func UseBlobHandlers(h BlobHandlerProvider) BlobOptionImpl {
	return &handler{h}
}

////////////////////////////////////////////////////////////////////////////////

// TargetElement described the index used to set the
// resource or source for the SetXXX calls.
// If -1 is returned an append is enforced.
type TargetElement interface {
	GetTargetIndex(resources compdesc.ElementListAccessor, meta accessors.ElementMeta) (int, error)
}

type TargetOptionImpl interface {
	TargetElementOption
	ModificationOption
	BlobModificationOption
}

type TargetElementOptions struct {
	TargetElement TargetElement

	// DisableExtraIdentityDefaulting disables the implicit defaulting of the extraIdentity.
	// A transfer operation must set this flag to preserve the normalizations.
	DisableExtraIdentityDefaulting *bool
}

type TargetElementOption interface {
	ApplyTargetOption(options *TargetElementOptions)
}

func (m *TargetElementOptions) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m *TargetElementOptions) ApplyModificationOption(opts *ModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m *TargetElementOptions) ApplyElementModificationOption(opts *ElementModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m *TargetElementOptions) ApplyTargetOption(opts *TargetElementOptions) {
	optionutils.Transfer(&opts.TargetElement, m.TargetElement)
	optionutils.Transfer(&opts.DisableExtraIdentityDefaulting, m.DisableExtraIdentityDefaulting)
}

func (m *TargetElementOptions) IsDisableExtraIdentityDefaulting() bool {
	return utils.AsBool(m.DisableExtraIdentityDefaulting)
}

func (m *TargetElementOptions) ApplyTargetOptions(list ...TargetElementOption) *TargetElementOptions {
	for _, o := range list {
		if o != nil {
			o.ApplyTargetOption(m)
		}
	}
	return m
}

func NewTargetElementOptions(list ...TargetElementOption) *TargetElementOptions {
	var m TargetElementOptions
	m.ApplyTargetOptions(list...)
	return &m
}

type ElementModificationOption interface {
	ApplyElementModificationOption(opts *ElementModificationOptions)
}

type ElementModificationOptions struct {
	TargetElementOptions

	// ModifyElement disables the modification of signature relevant
	// resource parts.
	ModifyElement *bool
}

func (m *ElementModificationOptions) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyElementModificationOption(&opts.ElementModificationOptions)
}

func (m *ElementModificationOptions) ApplyModificationOption(opts *ModificationOptions) {
	m.ApplyElementModificationOption(&opts.ElementModificationOptions)
}

func (m *ElementModificationOptions) ApplyElementModificationOption(opts *ElementModificationOptions) {
	optionutils.Transfer(&opts.ModifyElement, m.ModifyElement)
}

func (m *ElementModificationOptions) ApplyElementModificationOptions(list ...ElementModificationOption) *ElementModificationOptions {
	for _, o := range list {
		if o != nil {
			o.ApplyElementModificationOption(m)
		}
	}
	return m
}

func (m *ElementModificationOptions) IsModifyElement(def ...bool) bool {
	return utils.AsBool(m.ModifyElement, def...)
}

func NewElementModificationOptions(list ...ElementModificationOption) *ElementModificationOptions {
	var m ElementModificationOptions
	m.ApplyElementModificationOptions(list...)
	return &m
}

type ModificationOption interface {
	ApplyModificationOption(opts *ModificationOptions)
}

type ModOptionImpl interface {
	ModificationOption
	BlobModificationOption
}

type ElemModOptionImpl interface {
	ElementModificationOption
	ModificationOption
	BlobModificationOption
}

type ModificationOptions struct {
	ElementModificationOptions

	// AcceptExistentDigests don't validate/recalculate the content digest
	// of resources.
	AcceptExistentDigests *bool

	// DefaultHashAlgorithm is the hash algorithm to use if no specific setting os found
	DefaultHashAlgorithm string

	// HasherProvider is the factory for hash algorithms to use.
	HasherProvider HasherProvider

	// SkipVerify disabled the verification of given digests
	SkipVerify *bool

	// SkipDigest disabled digest creation (for legacy code, only!)
	SkipDigest *bool
}

func (m *ModificationOptions) IsAcceptExistentDigests() bool {
	return utils.AsBool(m.AcceptExistentDigests)
}

func (m *ModificationOptions) IsSkipDigest() bool {
	return utils.AsBool(m.SkipDigest)
}

func (m *ModificationOptions) IsSkipVerify() bool {
	return utils.AsBool(m.SkipVerify)
}

func (m *ModificationOptions) ApplyModificationOptions(list ...ModificationOption) *ModificationOptions {
	for _, o := range list {
		if o != nil {
			o.ApplyModificationOption(m)
		}
	}
	return m
}

func (m *ModificationOptions) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m *ModificationOptions) ApplyModificationOption(opts *ModificationOptions) {
	m.TargetElementOptions.ApplyTargetOption(&opts.TargetElementOptions)
	optionutils.Transfer(&opts.ModifyElement, m.ModifyElement)
	optionutils.Transfer(&opts.AcceptExistentDigests, m.AcceptExistentDigests)
	optionutils.Transfer(&opts.SkipDigest, m.SkipDigest)
	optionutils.Transfer(&opts.SkipVerify, m.SkipVerify)
	optionutils.Transfer(&opts.HasherProvider, m.HasherProvider)
	optionutils.Transfer(&opts.DefaultHashAlgorithm, m.DefaultHashAlgorithm)
}

func (m *ModificationOptions) GetHasher(algo ...string) Hasher {
	return m.HasherProvider.GetHasher(general.OptionalDefaulted(m.DefaultHashAlgorithm, algo...))
}

func NewModificationOptions(list ...ModificationOption) *ModificationOptions {
	var m ModificationOptions
	m.ApplyModificationOptions(list...)
	return &m
}

////////////////////////////////////////////////////////////////////////////////

type TargetIndex int

func (m TargetIndex) GetTargetIndex(elems compdesc.ElementListAccessor, meta accessors.ElementMeta) (int, error) {
	if int(m) >= elems.Len() {
		return -1, nil
	}
	return int(m), nil
}

func (m TargetIndex) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m TargetIndex) ApplyModificationOption(opts *ModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m TargetIndex) ApplyElementModificationOption(opts *ElementModificationOptions) {
	if m < 0 {
		opts.ModifyElement = generics.PointerTo(true)
	}
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m TargetIndex) ApplyTargetOption(opts *TargetElementOptions) {
	opts.TargetElement = m
}

////////////////////////////////////////////////////////////////////////////////

type TargetIdentityOrAppend v1.Identity

func (m TargetIdentityOrAppend) GetTargetIndex(elems compdesc.ElementListAccessor, meta accessors.ElementMeta) (int, error) {
	idx, _ := TargetIdentity(m).GetTargetIndex(elems, meta)
	return idx, nil
}

func (m TargetIdentityOrAppend) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m TargetIdentityOrAppend) ApplyModificationOption(opts *ModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m TargetIdentityOrAppend) ApplyElementModificationOption(opts *ElementModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m TargetIdentityOrAppend) ApplyTargetOption(opts *TargetElementOptions) {
	opts.TargetElement = m
}

////////////////////////////////////////////////////////////////////////////////

type TargetIdentity v1.Identity

func (m TargetIdentity) GetTargetIndex(elems compdesc.ElementListAccessor, meta accessors.ElementMeta) (int, error) {
	for i := 0; i < elems.Len(); i++ {
		r := elems.Get(i)
		if r.GetMeta().GetIdentity(elems).Equals(v1.Identity(m)) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("element %s not found", v1.Identity(m))
}

func (m TargetIdentity) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m TargetIdentity) ApplyModificationOption(opts *ModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m TargetIdentity) ApplyElementModificationOption(opts *ElementModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m TargetIdentity) ApplyTargetOption(opts *TargetElementOptions) {
	opts.TargetElement = m
}

////////////////////////////////////////////////////////////////////////////////

type disableextraidentitydefaulting bool

func (m disableextraidentitydefaulting) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m disableextraidentitydefaulting) ApplyModificationOption(opts *ModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m disableextraidentitydefaulting) ApplyElementModificationOption(opts *ElementModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m disableextraidentitydefaulting) ApplyTargetOption(opts *TargetElementOptions) {
	opts.DisableExtraIdentityDefaulting = utils.BoolP(m)
}

// DisableExtraIdentityDefaulting disables the defaulting of the extra identity.
func DisableExtraIdentityDefaulting(flag ...bool) TargetOptionImpl {
	return disableextraidentitydefaulting(utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

type replaceElement struct{}

var UpdateElement = replaceElement{}

func (m replaceElement) GetTargetIndex(elems compdesc.ElementListAccessor, meta accessors.ElementMeta) (int, error) {
	id := meta.GetIdentity(elems)
	for i := 0; i < elems.Len(); i++ {
		if elems.Get(i).GetMeta().GetIdentity(elems).Equals(id) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("element %s not found", id)
}

func (m replaceElement) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m replaceElement) ApplyModificationOption(opts *ModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m replaceElement) ApplyElementModificationOption(opts *ElementModificationOptions) {
	m.ApplyTargetOption(&opts.TargetElementOptions)
}

func (m replaceElement) ApplyTargetOption(opts *TargetElementOptions) {
	opts.TargetElement = m
}

////////////////////////////////////////////////////////////////////////////////

type modifyelement bool

func (m modifyelement) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m modifyelement) ApplyModificationOption(opts *ModificationOptions) {
	opts.ModifyElement = utils.BoolP(m)
}

func (m modifyelement) ApplyElementModificationOption(opts *ElementModificationOptions) {
	opts.ModifyElement = utils.BoolP(m)
}

func ModifyResource(flag ...bool) ModOptionImpl {
	return modifyelement(utils.OptionalDefaultedBool(true, flag...))
}

func ModifyElement(flag ...bool) ElemModOptionImpl {
	return modifyelement(utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

type acceptdigests bool

func (m acceptdigests) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m acceptdigests) ApplyModificationOption(opts *ModificationOptions) {
	opts.AcceptExistentDigests = utils.BoolP(m)
}

func AcceptExistentDigests(flag ...bool) ModOptionImpl {
	return acceptdigests(utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

type hashalgo string

func (m hashalgo) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m hashalgo) ApplyModificationOption(opts *ModificationOptions) {
	opts.DefaultHashAlgorithm = string(m)
}

func WithDefaultHashAlgorithm(algo ...string) ModOptionImpl {
	return hashalgo(utils.Optional(algo...))
}

////////////////////////////////////////////////////////////////////////////////

type hashprovider struct {
	prov HasherProvider
}

func (m hashprovider) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m *hashprovider) ApplyModificationOption(opts *ModificationOptions) {
	opts.HasherProvider = m.prov
}

func WithHasherProvider(prov HasherProvider) ModOptionImpl {
	return &hashprovider{prov}
}

////////////////////////////////////////////////////////////////////////////////

type skipverify bool

func (m skipverify) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m skipverify) ApplyModificationOption(opts *ModificationOptions) {
	opts.SkipVerify = utils.BoolP(m)
}

func SkipVerify(flag ...bool) ModOptionImpl {
	return skipverify(utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

type skipdigest bool

func (m skipdigest) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m skipdigest) ApplyModificationOption(opts *ModificationOptions) {
	opts.SkipDigest = utils.BoolP(m)
}

// SkipDigest disables digest creation if enabled.
//
// Deprecated: for legacy code, only.
func SkipDigest(flag ...bool) ModOptionImpl {
	return skipdigest(utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

// BlobModificationOption is used for option list allowing both,
// blob upload and modification options.
type BlobModificationOption interface {
	ApplyBlobModificationOption(*BlobModificationOptions)
}

type BlobModificationOptions struct {
	BlobUploadOptions
	ModificationOptions
}

func NewBlobModificationOptions(list ...BlobModificationOption) *BlobModificationOptions {
	var m BlobModificationOptions
	m.ApplyBlobModificationOptions(list...)
	return &m
}

func (m *BlobModificationOptions) ApplyBlobModificationOptions(list ...BlobModificationOption) {
	for _, o := range list {
		if o != nil {
			o.ApplyBlobModificationOption(m)
		}
	}
}

func (o *BlobModificationOptions) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	o.BlobUploadOptions.ApplyBlobUploadOption(&opts.BlobUploadOptions)
	o.ModificationOptions.ApplyModificationOption(&opts.ModificationOptions)
}

func (o *BlobModificationOptions) ApplyBlobUploadOption(opts *BlobUploadOptions) {
	o.BlobUploadOptions.ApplyBlobUploadOption(opts)
}

func (o *BlobModificationOptions) ApplyModificationOption(opts *ModificationOptions) {
	o.ModificationOptions.ApplyModificationOption(opts)
}

///////////////////////////////////////////////////////////////////////////////

// BlobModificationOption is used for option list allowing both,
// blob upload and modification options.
type AddVersionOption interface {
	ApplyAddVersionOption(*AddVersionOptions)
}

type AddVersionOptions struct {
	Overwrite *bool
	BlobUploadOptions
}

func NewAddVersionOptions(list ...AddVersionOption) *AddVersionOptions {
	var m AddVersionOptions
	m.ApplyAddVersionOptions(list...)
	return &m
}

func (m *AddVersionOptions) ApplyAddVersionOptions(list ...AddVersionOption) {
	for _, o := range list {
		if o != nil {
			o.ApplyAddVersionOption(m)
		}
	}
}

func (o *AddVersionOptions) ApplyAddVersionOption(opts *AddVersionOptions) {
	optionutils.ApplyOption(o.Overwrite, &opts.Overwrite)
	o.BlobUploadOptions.ApplyBlobUploadOption(&opts.BlobUploadOptions)
}

////////////////////////////////////////////////////////////////////////////////

type overwrite bool

func (m overwrite) ApplyAddVersionOption(opts *AddVersionOptions) {
	opts.Overwrite = utils.BoolP(m)
}

// Overwrite enabled the overwrite mode for adding a component version.
func Overwrite(flag ...bool) AddVersionOption {
	return overwrite(utils.OptionalDefaultedBool(true, flag...))
}
