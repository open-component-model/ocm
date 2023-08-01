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
}

func (m *ModificationOptions) Eval(list ...ModificationOption) {
	for _, o := range list {
		if o != nil {
			o.ApplyModificationOption(m)
		}
	}
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
	return modifyresource(utils.OptionalDefaultedBool(true, flag...))
}
