// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"golang.org/x/exp/slices"

	"github.com/open-component-model/ocm/pkg/runtime"
)

type decoder runtime.TypedObjectDecoder[runtime.TypedObject]

type AccessMethodBase struct {
	decoder
	nameDescription

	version string
	format  string
}

func MustNewAccessMethodBase(name, version string, proto AccessSpec, desc string, format string) AccessMethodBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}

	return AccessMethodBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		version: version,
		format:  format,
	}
}

func (b *AccessMethodBase) Version() string {
	return b.version
}

func (b *AccessMethodBase) Format() string {
	return b.format
}

////////////////////////////////////////////////////////////////////////////////

type UploaderBase = nameDescription

func MustNewUploaderBase(name, desc string) UploaderBase {
	return UploaderBase{
		name: name,
		desc: desc,
	}
}

////////////////////////////////////////////////////////////////////////////////

type ValueSetBase struct {
	decoder
	nameDescription

	version string
	format  string

	purposes []string
}

func MustNewValueSetBase(name, version string, proto runtime.TypedObject, purposes []string, desc string, format string) ValueSetBase {
	decoder, err := runtime.NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}
	return ValueSetBase{
		decoder: decoder,
		nameDescription: nameDescription{
			name: name,
			desc: desc,
		},
		version:  version,
		format:   format,
		purposes: slices.Clone(purposes),
	}
}

func (b *ValueSetBase) Version() string {
	return b.version
}

func (b *ValueSetBase) Format() string {
	return b.format
}

func (b *ValueSetBase) Purposes() []string {
	return b.purposes
}

////////////////////////////////////////////////////////////////////////////////

type nameDescription struct {
	name string
	desc string
}

func (b *nameDescription) Name() string {
	return b.name
}

func (b *nameDescription) Description() string {
	return b.desc
}

////////////////////////////////////////////////////////////////////////////////

// Config is a generic structured config stored in a string map.
type Config map[string]interface{}

func (c Config) GetValue(name string) (interface{}, bool) {
	v, ok := c[name]
	return v, ok
}
