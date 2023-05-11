// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package scheme

import (
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Converter[O Object] interface {
	runtime.Converter[O] // Go live, Converter[O] and runtime.Converter[O] are not type compatible!
}

type defaultType[O Object] struct {
	runtime.VersionedTypedObjectType
}

func NewTypeByProtoType[O Object](proto Object, converter Converter[O]) Type[O] {
	return &defaultType[O]{runtime.NewVersionedTypedObjectTypeByConverter("", proto, runtime.Converter[O](converter))} // wow, I love Go
}

func (d *defaultType[O]) Decode(data []byte, unmarshaler runtime.Unmarshaler) (O, error) {
	var zero O
	o, err := d.VersionedTypedObjectType.Decode(data, unmarshaler)
	if err != nil {
		return zero, err
	}
	return o.(O), nil
}

func (d *defaultType[O]) Encode(o O, marshaler runtime.Marshaler) ([]byte, error) {
	return d.VersionedTypedObjectType.Encode(o, marshaler)
}

////////////////////////////////////////////////////////////////////////////////

func NewIdentityType[O Object](proto O) Type[O] {
	return NewTypeByProtoType[O](proto, runtime.IdentityConverter[O]{})
}
