// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"reflect"
)

// TypedObjectConverter converts a versioned representation into the
// intended type required by the scheme.
type TypedObjectConverter[T TypedObject] interface {
	ConvertTo(in interface{}) (T, error)
}

// ConvertingDecoder uses a serialization form different from the
// intended object type, that is converted to achieve the decode result.
type ConvertingDecoder[T TypedObject] struct {
	proto reflect.Type
	TypedObjectConverter[T]
}

var _ TypedObjectDecoder[TypedObject] = &ConvertingDecoder[TypedObject]{}

func MustNewConvertingDecoder[T TypedObject](proto interface{}, conv TypedObjectConverter[T]) *ConvertingDecoder[T] {
	d, err := NewConvertingDecoder[T](proto, conv)
	if err != nil {
		panic(err)
	}
	return d
}

func NewConvertingDecoder[T TypedObject](proto interface{}, conv TypedObjectConverter[T]) (*ConvertingDecoder[T], error) {
	t, err := ProtoType(proto)
	if err != nil {
		return nil, err
	}
	return &ConvertingDecoder[T]{
		proto:                t,
		TypedObjectConverter: conv,
	}, nil
}

func (d *ConvertingDecoder[T]) Decode(data []byte, unmarshaler Unmarshaler) (T, error) {
	var zero T

	versioned := d.CreateData()
	err := unmarshaler.Unmarshal(data, versioned)
	if err != nil {
		return zero, err
	}
	return d.ConvertTo(versioned)
}

func (d *ConvertingDecoder[T]) CreateData() interface{} {
	return reflect.New(d.proto).Interface()
}

////////////////////////////////////////////////////////////////////////////////

type Converter[T TypedObject] interface {
	ConvertFrom(object T) (TypedObject, error)
	ConvertTo(object interface{}) (T, error)
}

func NewConvertWrapper[T TypedObject](c Converter[T]) TypedObjectConverter[T] {
	return c
}

////////////////////////////////////////////////////////////////////////////////

type FormatVersion[T TypedObject] interface {
	Converter[T]
	TypedObjectDecoder[T]
	CreateData() interface{}
}

////////////////////////////////////////////////////////////////////////////////

type formatVersion[T VersionedTypedObject] struct {
	*ConvertingDecoder[T]
	Converter[T]
}

func NewProtoBasedVersion[T VersionedTypedObject](proto VersionedTypedObject, converter Converter[T]) FormatVersion[T] {
	return &formatVersion[T]{
		ConvertingDecoder: MustNewConvertingDecoder(proto, NewConvertWrapper(converter)), // why does using converter directly not work?
		Converter:         converter,
	}
}

// ConvertedType if the interface of an object type for versioned objects with conversion.
type ConvertedType[T VersionedTypedObject] interface {
	FormatVersion[T]
	VersionedTypeInfo
	Encode(obj T, m Marshaler) ([]byte, error)
}

// ObjectConvertedType is a default implementation for a ConvertedType.
type ObjectConvertedType[T VersionedTypedObject] struct {
	FormatVersion[T]
	VersionedTypeInfo
}

var (
	_ FormatVersion[VersionedTypedObject]      = &ObjectConvertedType[VersionedTypedObject]{}
	_ TypedObjectEncoder[VersionedTypedObject] = &ObjectConvertedType[VersionedTypedObject]{}
)

func NewConvertedType[T VersionedTypedObject](name string, v FormatVersion[T]) *ObjectConvertedType[T] {
	t := &ObjectConvertedType[T]{
		VersionedTypeInfo: NewVersionedObjectType(name),
		FormatVersion:     v,
	}
	return t
}

func (t *ObjectConvertedType[T]) Encode(obj T, m Marshaler) ([]byte, error) {
	c, err := t.ConvertFrom(obj)
	if err != nil {
		return nil, err
	}
	return m.Marshal(c)
}
