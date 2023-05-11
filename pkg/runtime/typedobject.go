// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

// ObjectTypedObject is the minimal implementation of a typed object
// managing the type information.
type ObjectTypedObject = ObjectType

func NewTypedObject(typ string) ObjectTypedObject {
	return NewObjectType(typ)
}

// TypedObject defines the common interface for all kinds of typed objects.
type TypedObject interface {
	TypeInfo
}

// TypedObjectType is the interface for a type object for an TypedObject.
type TypedObjectType interface {
	TypeInfo
	TypedObjectDecoder
}

type typeObject struct {
	_ObjectType
	_TypedObjectDecoder
}

var _ TypedObjectType = (*typeObject)(nil)

func NewTypedObjectTypeByDecoder(name string, decoder TypedObjectDecoder) TypedObjectType {
	return &typeObject{
		_ObjectType:         NewObjectType(name),
		_TypedObjectDecoder: decoder,
	}
}

func NewTypedObjectTypeByProto(name string, proto TypedObject) TypedObjectType {
	return &typeObject{
		_ObjectType:         NewObjectType(name),
		_TypedObjectDecoder: MustNewDirectDecoder(proto),
	}
}
