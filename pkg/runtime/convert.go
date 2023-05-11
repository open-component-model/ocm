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
		TypedObjectDecoder
		TypedObjectEncoder
	}
	// _FormatVersion[T TypedObject] = FormatVersion[T] // I like Go.
	_FormatVersion[T TypedObject] interface {
		FormatVersion[T]
	}
)

type formatVersion[T VersionedTypedObject] struct {
	decoder   TypedObjectDecoder
	converter Converter[T]
}

func (c *formatVersion[T]) Encode(object TypedObject, marshaler Marshaler) ([]byte, error) {
	v, err := c.converter.ConvertFrom(object.(T))
	if err != nil {
		return nil, err
	}
	if marshaler == nil {
		marshaler = DefaultJSONEncoding
	}
	return marshaler.Marshal(v)
}

func (c *formatVersion[T]) Decode(data []byte, unmarshaler Unmarshaler) (TypedObject, error) {
	v, err := c.decoder.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return c.converter.ConvertTo(v)
}

func NewProtoBasedVersion[T VersionedTypedObject](proto VersionedTypedObject, converter Converter[T]) FormatVersion[T] {
	return &formatVersion[T]{
		decoder:   MustNewDirectDecoder(proto),
		converter: converter,
	}
}
