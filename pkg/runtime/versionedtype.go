// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

const VersionSeparator = "/"

// VersionedTypeInfo in the accessor for versioned type information.
type VersionedTypeInfo interface {
	TypeInfo
	GetKind() string
	GetVersion() string
}

// VersionedTypedObject in an instance of a VersionedTypedObjectType.
type VersionedTypedObject interface {
	TypedObject
	VersionedTypeInfo
}

////////////////////////////////////////////////////////////////////////////////
// Versioned Typed Objects

// ObjectVersionedTypedObject is a minimal implementation of a VersionedTypedObject.
type ObjectVersionedTypedObject = VersionedObjectType

// ObjectVersionedType is a minimal implementation of a VersionedTypedObject.
// For compatibility, we keep the old not aligned type name.
type ObjectVersionedType = ObjectVersionedTypedObject

// NewVersionedTypedObject creates an ObjectVersionedType value.
func NewVersionedTypedObject(args ...string) ObjectVersionedTypedObject {
	return ObjectVersionedTypedObject{Type: TypeName(args...)}
}

////////////////////////////////////////////////////////////////////////////////
// Object Types for Versioned Typed Objects

// InternalVersionedTypedObject is the base type used
// by *internal* representations of versioned specification
// formats. It is used to convert from/to dedicated
// format versions.
type InternalVersionedTypedObject struct {
	ObjectVersionedType
	encoder TypedObjectEncoder
}

var _ encoder = (*InternalVersionedTypedObject)(nil)

type encoder interface {
	encode(obj VersionedTypedObject) ([]byte, error)
}

func NewInternalVersionedTypedObject(encoder TypedObjectEncoder, types ...string) InternalVersionedTypedObject {
	return InternalVersionedTypedObject{
		ObjectVersionedType: NewVersionedTypedObject(types...),
		encoder:             encoder,
	}
}

func (o *InternalVersionedTypedObject) encode(obj VersionedTypedObject) ([]byte, error) {
	return o.encoder.Encode(obj, DefaultJSONEncoding)
}

func GetEncoder(obj VersionedTypedObject) encoder {
	if e, ok := obj.(encoder); ok {
		return e
	}
	return nil
}

func MarshalVersionedTypedObject(obj VersionedTypedObject, toe ...TypedObjectEncoder) ([]byte, error) {
	if e := GetEncoder(obj); e != nil {
		return e.encode(obj)
	}
	e := utils.Optional(toe...)
	if e != nil {
		return e.Encode(obj, DefaultJSONEncoding)
	}
	return nil, errors.ErrUnknown("object type", obj.GetType())
}

////////////////////////////////////////////////////////////////////////////////

// VersionedTypedObjectType is the interface of a type object for a versioned type.
type VersionedTypedObjectType interface {
	VersionedTypeInfo
	TypedObjectDecoder
	TypedObjectEncoder
}

type versionedTypedObjectType struct {
	_VersionedObjectType
	_FormatVersion[VersionedObjectType]
}

var _ FormatVersion[VersionedTypedObject] = (*versionedTypedObjectType)(nil)

func NewVersionedTypedObjectTypeByProto[T VersionedTypedObject](name string, proto T) VersionedTypedObjectType {
	return &versionedTypedObjectType{
		_VersionedObjectType: NewVersionedObjectType(name),
		_FormatVersion:       NewProtoBasedVersion[T](proto, IdentityConverter[T]{}),
	}
}

func NewVersionedTypedObjectTypeByConverter[T VersionedTypedObject](name string, proto VersionedTypedObject, converter Converter[T]) VersionedTypedObjectType {
	return &versionedTypedObjectType{
		_VersionedObjectType: NewVersionedObjectType(name),
		_FormatVersion:       NewProtoBasedVersion[T](proto, converter),
	}
}

func NewVersionedTypedObjectTypeByVersion[T VersionedTypedObject](name string, version FormatVersion[T]) VersionedTypedObjectType {
	return &versionedTypedObjectType{
		_VersionedObjectType: NewVersionedObjectType(name),
		_FormatVersion:       version,
	}
}

////////////////////////////////////////////////////////////////////////////////

func TypeName(args ...string) string {
	if len(args) == 1 {
		return args[0]
	}
	if len(args) == 2 {
		if args[1] == "" {
			return args[0]
		}
		return args[0] + VersionSeparator + args[1]
	}
	panic("invalid call to TypeName, one or two arguments required")
}

func KindVersion(t string) (string, string) {
	i := strings.LastIndex(t, VersionSeparator)
	if i > 0 {
		return t[:i], t[i+1:]
	}
	return t, ""
}

func GetKind(v TypedObject) string {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		return t
	}
	return t[:i]
}

func GetVersion(v TypedObject) string {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		return "v1"
	}
	return t[i+1:]
}
