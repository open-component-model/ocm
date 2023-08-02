// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"github.com/open-component-model/ocm/pkg/utils"
)

type ModificationOption interface {
	ApplyModificationOption(opts *ModificationOptions)
}

type ModificationOptions struct {
	// ModifyResource disables the modification of signature releveant
	// resource parts.
	ModifyResource bool

	// AcceptExistentDigests don't validate/recalculate the content digest
	// of resources.
	AcceptExistentDigests bool

	// DefaultHashAlgorithm is the hash algorithm to use if no specific setting os found
	DefaultHashAlgorithm string

	// HasherProvider is the factory for hash algorithms to use.
	HasherProvider HasherProvider
}

func (m *ModificationOptions) Eval(list ...ModificationOption) {
	for _, o := range list {
		if o != nil {
			o.ApplyModificationOption(m)
		}
	}
}

func (m *ModificationOptions) GetHasher(algo ...string) Hasher {
	return m.HasherProvider.GetHasher(utils.OptionalDefaulted(m.DefaultHashAlgorithm, algo...))
}

func EvalModificationOptions(list ...ModificationOption) ModificationOptions {
	var m ModificationOptions
	m.Eval(list...)
	return m
}

////////////////////////////////////////////////////////////////////////////////

type modifyresource bool

func (m modifyresource) ApplyModificationOption(opts *ModificationOptions) {
	opts.ModifyResource = bool(m)
}

func ModifyResource(flag ...bool) ModificationOption {
	return modifyresource(utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

type acceptdigests bool

func (m acceptdigests) ApplyModificationOption(opts *ModificationOptions) {
	opts.AcceptExistentDigests = bool(m)
}

func AcceptExistentDigests(flag ...bool) ModificationOption {
	return acceptdigests(utils.OptionalDefaultedBool(true, flag...))
}

////////////////////////////////////////////////////////////////////////////////

type hashalgo string

func (m hashalgo) ApplyModificationOption(opts *ModificationOptions) {
	opts.DefaultHashAlgorithm = string(m)
}

func WithDefaultHashAlgorithm(algo ...string) ModificationOption {
	return hashalgo(utils.Optional(algo...))
}

////////////////////////////////////////////////////////////////////////////////

type hashprovider struct {
	prov HasherProvider
}

func (m *hashprovider) ApplyModificationOption(opts *ModificationOptions) {
	opts.HasherProvider = m.prov
}

func WithHasherProvider(prov HasherProvider) ModificationOption {
	return &hashprovider{prov}
}
