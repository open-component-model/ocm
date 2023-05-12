// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package scheme

import (
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Converter[O Object] interface {
	runtime.Converter[O, runtime.TypedObject] // Go live, Converter[O] and runtime.Converter[O] are not type compatible!
}

type defaultType[O Object] struct {
	runtime.VersionedTypedObjectType[O]
}

func NewTypeByProtoType[O Object](proto Object, converter Converter[O]) Type[O] {
	return &defaultType[O]{runtime.NewVersionedTypedObjectTypeByProtoConverter[O]("", proto, runtime.Converter[O, runtime.TypedObject](converter))} // wow, I love Go
}

func (d *defaultType[O]) Decode(data []byte, unmarshaler runtime.Unmarshaler) (O, error) {
	var zero O
	o, err := d.VersionedTypedObjectType.Decode(data, unmarshaler)
	if err != nil {
		return zero, err
	}
	return o, nil
}

func (d *defaultType[O]) Encode(o O, marshaler runtime.Marshaler) ([]byte, error) {
	return d.VersionedTypedObjectType.Encode(o, marshaler)
}

////////////////////////////////////////////////////////////////////////////////

func NewIdentityType[O Object](proto O) Type[O] {
	return NewTypeByProtoType[O](proto, IdentityConverter[O]{})
}

type IdentityConverter[O Object] struct{}

func (i IdentityConverter[O]) ConvertFrom(object O) (runtime.TypedObject, error) {
	return runtime.Cast[O, runtime.TypedObject](object)
}

func (i IdentityConverter[O]) ConvertTo(object runtime.TypedObject) (O, error) {
	return runtime.Cast[runtime.TypedObject, O](object)
}
