package ocm

import (
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/attrs/hashattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
)

type (
	TargetElement        = internal.TargetElement
	TargetElementOption  = internal.TargetElementOption
	TargetElementOptions = internal.TargetElementOptions

	ElementModificationOption  = internal.ElementModificationOption
	ElementModificationOptions = internal.ElementModificationOptions

	ModificationOption  = internal.ModificationOption
	ModificationOptions = internal.ModificationOptions

	BlobModificationOption  = internal.BlobModificationOption
	BlobModificationOptions = internal.BlobModificationOptions

	BlobUploadOption  = internal.BlobUploadOption
	BlobUploadOptions = internal.BlobUploadOptions

	AddVersionOption  = internal.AddVersionOption
	AddVersionOptions = internal.AddVersionOptions
)

////////////////////////////////////////////////////////////////////////////////

func NewAddVersionOptions(list ...AddVersionOption) *AddVersionOptions {
	return internal.NewAddVersionOptions(list...)
}

// Overwrite enabled the overwrite mode for adding a component version.
func Overwrite(flag ...bool) AddVersionOption {
	return internal.Overwrite(flag...)
}

////////////////////////////////////////////////////////////////////////////////

func NewBlobModificationOptions(list ...BlobModificationOption) *BlobModificationOptions {
	return internal.NewBlobModificationOptions(list...)
}

////////////////////////////////////////////////////////////////////////////////

func NewBlobUploadOptions(list ...BlobUploadOption) *BlobUploadOptions {
	return internal.NewBlobUploadOptions(list...)
}

func UseBlobHandlers(h BlobHandlerProvider) internal.BlobOptionImpl {
	return internal.UseBlobHandlers(h)
}

////////////////////////////////////////////////////////////////////////////////

func NewModificationOptions(list ...ModificationOption) *ModificationOptions {
	return internal.NewModificationOptions(list...)
}

func TargetIndex(idx int) internal.TargetOptionImpl {
	return internal.TargetIndex(idx)
}

// RawIdentity disables the defaulting of the extra identity
func RawIdentity(flag ...bool) internal.TargetOptionImpl {
	return internal.RawIdentity(flag...)
}

const AppendElement = internal.TargetIndex(-1)

var UpdateElement = internal.UpdateElement

func TargetIdentity(id v1.Identity) internal.TargetOptionImpl {
	return internal.TargetIdentity(id)
}

func TargetIdentityOrCreate(id v1.Identity) internal.TargetOptionImpl {
	return internal.TargetIdentityOrAppend(id)
}

// Deprecated: use ModifyElement.
func ModifyResource(flag ...bool) internal.ModOptionImpl {
	return internal.ModifyResource(flag...)
}

func ModifyElement(flag ...bool) internal.ElemModOptionImpl {
	return internal.ModifyElement(flag...)
}

func AcceptExistentDigests(flag ...bool) internal.ModOptionImpl {
	return internal.AcceptExistentDigests(flag...)
}

func WithDefaultHashAlgorithm(algo ...string) internal.ModOptionImpl {
	return internal.WithDefaultHashAlgorithm(algo...)
}

func WithHasherProvider(prov HasherProvider) internal.ModOptionImpl {
	return internal.WithHasherProvider(prov)
}

func SkipVerify(flag ...bool) internal.ModOptionImpl {
	return internal.SkipVerify(flag...)
}

// SkipDigest disables digest creation if enabled.
//
// Deprecated: for legacy code, only.
func SkipDigest(flag ...bool) internal.ModOptionImpl {
	return internal.SkipDigest(flag...)
}

///////////////////////////////////////////////////////

func CompleteModificationOptions(ctx ContextProvider, m *ModificationOptions) {
	attr := hashattr.Get(ctx.OCMContext())
	if m.DefaultHashAlgorithm == "" {
		m.DefaultHashAlgorithm = attr.DefaultHasher
	}
	if m.DefaultHashAlgorithm == "" {
		m.DefaultHashAlgorithm = sha256.Algorithm
	}
	if m.HasherProvider == nil {
		m.HasherProvider = signingattr.Get(ctx.OCMContext())
	}
}
