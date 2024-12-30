package cpi

import (
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/attrs/hashattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
)

type (
	TargetElement = internal.TargetElement
	TargetOption  = internal.TargetOption
	TargetOptions = internal.TargetOptions

	ModificationOption  = internal.ModificationOption
	ModificationOptions = internal.ModificationOptions

	BlobModificationOption  = internal.BlobModificationOption
	BlobModificationOptions = internal.BlobModificationOptions

	BlobUploadOption  = internal.BlobUploadOption
	BlobUploadOptions = internal.BlobUploadOptions

	AddVersionOption  = internal.AddVersionOption
	AddVersionOptions = internal.AddVersionOptions
)

func NewTargetOptions(list ...TargetOption) *TargetOptions {
	var m TargetOptions
	m.ApplyTargetOptions(list...)
	return &m
}

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

func TargetIndex(idx int) internal.TargetIndex {
	return internal.TargetIndex(-1)
}

const AppendElement = internal.TargetIndex(-1)

var UpdateElement = internal.UpdateElement

func TargetIdentity(id v1.Identity) internal.TargetIdentity {
	return internal.TargetIdentity(id)
}

func ModifyResource(flag ...bool) internal.ModOptionImpl {
	return internal.ModifyResource(flag...)
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
