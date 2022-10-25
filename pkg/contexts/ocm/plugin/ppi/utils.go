// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"github.com/open-component-model/ocm/pkg/runtime"
)

type decoder runtime.TypedObjectDecoder

type AccessMethodBase struct {
	decoder
	name    string
	version string
}

func MustNewAccessMethodBase(name, version string, proto AccessSpec) AccessMethodBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}

	return AccessMethodBase{
		decoder: decoder,
		name:    name,
		version: version,
	}
}

func (b *AccessMethodBase) Name() string {
	return b.name
}

func (b *AccessMethodBase) Version() string {
	return b.version
}
