// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"github.com/open-component-model/ocm/pkg/utils"
)

type BlobUploadOption interface {
	ApplyBlobUploadOption(opts *BlobUploadOptions)
}

type BlobOptionImpl interface {
	BlobUploadOption
	BlobModificationOption
}

type BlobUploadOptions struct {
	UseNoDefaultIfNotSet bool
	BlobHandlerProvider  BlobHandlerProvider
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
	if o.BlobHandlerProvider != nil {
		opts.BlobHandlerProvider = o.BlobHandlerProvider
	} else {
		opts.UseNoDefaultIfNotSet = true
	}
}

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

type ModificationOption interface {
	ApplyModificationOption(opts *ModificationOptions)
}

type ModOptionImpl interface {
	ModificationOption
	BlobModificationOption
}

type ModificationOptions struct {
	// ModifyResource disables the modification of signature releveant
	// resource parts.
	ModifyResource *bool

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

func (m *ModificationOptions) IsModifyResource() bool {
	return utils.AsBool(m.ModifyResource)
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
	applyBool(m.ModifyResource, &opts.ModifyResource)
	applyBool(m.AcceptExistentDigests, &opts.AcceptExistentDigests)
	applyBool(m.SkipDigest, &opts.SkipDigest)
	applyBool(m.SkipVerify, &opts.SkipVerify)
	if m.HasherProvider != nil {
		opts.HasherProvider = m.HasherProvider
	}
	if m.DefaultHashAlgorithm != "" {
		opts.DefaultHashAlgorithm = m.DefaultHashAlgorithm
	}
}

func applyBool(m *bool, t **bool) {
	if m != nil {
		*t = utils.BoolP(*m)
	}
}

func (m *ModificationOptions) GetHasher(algo ...string) Hasher {
	return m.HasherProvider.GetHasher(utils.OptionalDefaulted(m.DefaultHashAlgorithm, algo...))
}

func NewModificationOptions(list ...ModificationOption) *ModificationOptions {
	var m ModificationOptions
	m.ApplyModificationOptions(list...)
	return &m
}

////////////////////////////////////////////////////////////////////////////////

type modifyresource bool

func (m modifyresource) ApplyBlobModificationOption(opts *BlobModificationOptions) {
	m.ApplyModificationOption(&opts.ModificationOptions)
}

func (m modifyresource) ApplyModificationOption(opts *ModificationOptions) {
	opts.ModifyResource = utils.BoolP(m)
}

func ModifyResource(flag ...bool) ModOptionImpl {
	return modifyresource(utils.OptionalDefaultedBool(true, flag...))
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
