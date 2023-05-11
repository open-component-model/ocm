// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/errors"
)

type Converter[T TypedObject] interface {
	// ConvertFrom converts from an internal version into an external format.
	ConvertFrom(object T) (TypedObject, error)
	// ConvertTo converts from an external format into an internal version.
	ConvertTo(object interface{}) (T, error)
}

type IdentityConverter[T TypedObject] struct{}

var _ Converter[TypedObject] = (*IdentityConverter[TypedObject])(nil)

func (_ IdentityConverter[T]) ConvertFrom(object T) (TypedObject, error) {
	return object, nil
}

func (_ IdentityConverter[T]) ConvertTo(object interface{}) (T, error) {
	var zero T
	if t, ok := object.(T); ok {
		return t, nil
	}
	return zero, errors.ErrInvalid("type", fmt.Sprintf("%T", object))
}

////////////////////////////////////////////////////////////////////////////////

type (
	FormatVersion[T TypedObject] interface {
		TypedObjectDecoder[T]
		TypedObjectEncoder[T]
	}
	// _FormatVersion[T TypedObject] = FormatVersion[T] // I like Go.
	_FormatVersion[T TypedObject] interface {
		FormatVersion[T]
	}
)

type formatVersion[T VersionedTypedObject] struct {
	decoder   TypedObjectDecoder[TypedObject]
	converter Converter[T]
}

func (c *formatVersion[T]) Encode(object T, marshaler Marshaler) ([]byte, error) {
	v, err := c.converter.ConvertFrom(object)
	if err != nil {
		return nil, err
	}
	if marshaler == nil {
		marshaler = DefaultJSONEncoding
	}
	return marshaler.Marshal(v)
}

func (c *formatVersion[T]) Decode(data []byte, unmarshaler Unmarshaler) (T, error) {
	var zero T
	v, err := c.decoder.Decode(data, unmarshaler)
	if err != nil {
		return zero, err
	}
	return c.converter.ConvertTo(v)
}

// caster applies an implemantation to interface upcast for a format version,
// here I has to be a subtype of T, but thanks to Go this cannot be expressed.
type caster[T VersionedTypedObject, I VersionedTypedObject] struct {
	version FormatVersion[I]
}

func (c *caster[T, I]) Decode(data []byte, unmarshaler Unmarshaler) (T, error) {
	var zero T
	o, err := c.version.Decode(data, unmarshaler)
	if err != nil {
		return zero, err
	}
	var i interface{} = o // type parameter based casts not supported by go
	if t, ok := i.(T); ok {
		return t, nil
	}
	return zero, errors.ErrInvalid("type", fmt.Sprintf("%T", o))
}

func (c *caster[T, I]) Encode(o T, marshaler Marshaler) ([]byte, error) {
	var t interface{} = o // type parameter based casts not supported by go
	if i, ok := t.(I); ok {
		return c.version.Encode(i, marshaler)
	}
	return nil, errors.ErrInvalid("type", fmt.Sprintf("%T", o))
}

type implementation struct {
	VersionedTypedObject
}

var _ FormatVersion[VersionedTypedObject] = (*caster[VersionedTypedObject, implementation])(nil)

// NewProtoBasedVersion creates a new format version for versioned typed objects,
// where T is the common *interface* of all types of the same type realm and I is the
// *internal implementation* commonly used for the various version variants of a dedicated kind of type,
// representing the format this format version is responsible for.
// Therefore, I must be subtype of T, which cannot be expressed in Go.
// The converter must convert between the external version, specified by the given prototype and
// the *internal* representation (type I) used to internally represent a set of variants as Go object.
func NewProtoBasedVersion[T VersionedTypedObject, I VersionedTypedObject](proto VersionedTypedObject, converter Converter[I]) FormatVersion[T] {
	return &caster[T, I]{&formatVersion[I]{
		decoder:   MustNewDirectDecoder[TypedObject](proto),
		converter: converter,
	}}
}
