// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

type Converter[T TypedObject] interface {
	ConvertFrom(object T) (TypedObject, error)
	ConvertTo(object interface{}) (T, error)
}

// ConvertWrapper wraps a versioned typed converter into a generic
// converter to fulfill the more generic interface.
type ConvertWrapper[T VersionedTypedObject] struct {
	converter Converter[T]
}

func (c *ConvertWrapper[T]) ConvertTo(object interface{}) (TypedObject, error) {
	return c.converter.ConvertTo(object)
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

var _ TypedObject = (VersionedTypedObject)(nil)

func NewProtoBasedVersion[T VersionedTypedObject](proto VersionedTypedObject, converter Converter[T]) FormatVersion[T] {
	return &formatVersion[T]{
		ConvertingDecoder: MustNewConvertingDecoder(proto, NewConvertWrapper(converter)), // why does using converter directly not work?
		Converter:         converter,
	}
}

type ConvertedType[T VersionedTypedObject] struct {
	FormatVersion[T]
	VersionedType[T]
}

var (
	_ FormatVersion[VersionedTypedObject]      = &ConvertedType[VersionedTypedObject]{}
	_ TypedObjectEncoder[VersionedTypedObject] = &ConvertedType[VersionedTypedObject]{}
)

func NewConvertedType[T VersionedTypedObject](name string, v FormatVersion[T]) *ConvertedType[T] {
	t := &ConvertedType[T]{
		VersionedType: versionedType[T]{
			ObjectVersionedType: NewVersionedObjectType(name),
			TypedObjectDecoder:  v,
		},
		FormatVersion: v,
	}
	return t
}

func (t *ConvertedType[T]) Decode(data []byte, unmarshaler Unmarshaler) (T, error) {
	// resolve method resolution conflict, basically both candidates are identical.
	return t.VersionedType.Decode(data, unmarshaler)
}

func (t *ConvertedType[T]) Encode(obj T, m Marshaler) ([]byte, error) {
	c, err := t.ConvertFrom(obj)
	if err != nil {
		return nil, err
	}
	return m.Marshal(c)
}
