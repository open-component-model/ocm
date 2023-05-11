// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"reflect"
)

// TypedObjectConverter converts a versioned representation into the
// intended type required by the scheme.
type TypedObjectConverter interface {
	ConvertTo(in interface{}) (TypedObject, error)
}

// ConvertingDecoder uses a serialization form different from the
// intended object type, that is converted to achieve the decode result.
type ConvertingDecoder struct {
	proto reflect.Type
	TypedObjectConverter
}

var _ TypedObjectDecoder = &ConvertingDecoder{}

func MustNewConvertingDecoder(proto interface{}, conv TypedObjectConverter) *ConvertingDecoder {
	d, err := NewConvertingDecoder(proto, conv)
	if err != nil {
		panic(err)
	}
	return d
}

func NewConvertingDecoder(proto interface{}, conv TypedObjectConverter) (*ConvertingDecoder, error) {
	t, err := ProtoType(proto)
	if err != nil {
		return nil, err
	}
	return &ConvertingDecoder{
		proto:                t,
		TypedObjectConverter: conv,
	}, nil
}

func (d *ConvertingDecoder) Decode(data []byte, unmarshaler Unmarshaler) (TypedObject, error) {
	versioned := d.CreateData()
	err := unmarshaler.Unmarshal(data, versioned)
	if err != nil {
		return nil, err
	}
	return d.ConvertTo(versioned)
}

func (d *ConvertingDecoder) CreateData() interface{} {
	return reflect.New(d.proto).Interface()
}

////////////////////////////////////////////////////////////////////////////////

type Converter[T TypedObject] interface {
	ConvertFrom(object T) (TypedObject, error)
	ConvertTo(object interface{}) (T, error)
}

// ConvertWrapper wraps a versioned typed converter into a generic
// converter to fulfill the more generic interface.
type ConvertWrapper[T TypedObject] struct {
	converter Converter[T]
}

func (c *ConvertWrapper[T]) ConvertTo(object interface{}) (TypedObject, error) {
	return c.converter.ConvertTo(object)
}

func NewConvertWrapper[T TypedObject](c Converter[T]) TypedObjectConverter {
	return &ConvertWrapper[T]{converter: c}
}

////////////////////////////////////////////////////////////////////////////////

type FormatVersion[T TypedObject] interface {
	Converter[T]
	TypedObjectDecoder
	CreateData() interface{}
}

////////////////////////////////////////////////////////////////////////////////

type formatVersion[T VersionedTypedObject] struct {
	*ConvertingDecoder
	Converter[T]
}

func NewProtoBasedVersion[T VersionedTypedObject](proto VersionedTypedObject, converter Converter[T]) FormatVersion[T] {
	return &formatVersion[T]{
		ConvertingDecoder: MustNewConvertingDecoder(proto, NewConvertWrapper(converter)),
		Converter:         converter,
	}
}

// ConvertedType if the interface of an object type for versioned objects with conversion.
type ConvertedType[T VersionedTypedObject] interface {
	FormatVersion[T]
	VersionedTypeInfo
	Encode(obj TypedObject, m Marshaler) ([]byte, error)
}

// ObjectConvertedType is a default implementation for a ConvertedType.
type ObjectConvertedType[T VersionedTypedObject] struct {
	FormatVersion[T]
	VersionedTypeInfo
}

var (
	_ FormatVersion[VersionedTypedObject] = &ObjectConvertedType[VersionedTypedObject]{}
	_ TypedObjectEncoder                  = &ObjectConvertedType[VersionedTypedObject]{}
)

func NewConvertedType[T VersionedTypedObject](name string, v FormatVersion[T]) *ObjectConvertedType[T] {
	t := &ObjectConvertedType[T]{
		VersionedTypeInfo: NewVersionedObjectType(name),
		FormatVersion:     v,
	}
	return t
}

func (t *ObjectConvertedType[T]) Encode(obj TypedObject, m Marshaler) ([]byte, error) {
	c, err := t.ConvertFrom(obj.(T))
	if err != nil {
		return nil, err
	}
	return m.Marshal(c)
}
