// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"reflect"
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

const VersionSeparator = "/"

type VersionedTypedObject interface {
	TypedObject
	GetKind() string
	GetVersion() string
}

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

type ObjectVersionedType ObjectType

// NewVersionedObjectType creates an ObjectVersionedType value.
func NewVersionedObjectType(args ...string) ObjectVersionedType {
	return ObjectVersionedType{Type: TypeName(args...)}
}

// GetType returns the type of the object.
func (t ObjectVersionedType) GetType() string {
	return t.Type
}

// SetType sets the type of the object.
func (t *ObjectVersionedType) SetType(typ string) {
	t.Type = typ
}

// GetKind returns the kind of the object.
func (v ObjectVersionedType) GetKind() string {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		return t
	}
	return t[:i]
}

// SetKind sets the kind of the object.
func (v *ObjectVersionedType) SetKind(kind string) {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		v.Type = kind
	} else {
		v.Type = kind + t[i:]
	}
}

// GetVersion returns the version of the object.
func (v ObjectVersionedType) GetVersion() string {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		return "v1"
	}
	return t[i+1:]
}

// SetVersion sets the version of the object.
func (v *ObjectVersionedType) SetVersion(version string) {
	t := v.GetType()
	i := strings.LastIndex(t, VersionSeparator)
	if i < 0 {
		if version != "" {
			v.Type = v.Type + VersionSeparator + version
		}
	} else {
		if version != "" {
			v.Type = t[:i] + VersionSeparator + version
		} else {
			v.Type = t[:i]
		}
	}
}

type InternalVersionedObjectType[T VersionedTypedObject] struct {
	ObjectVersionedType
	encoder TypedObjectEncoder[T]
}

var _ encoder = (*InternalVersionedObjectType[VersionedTypedObject])(nil)

type encoder interface {
	encode(obj VersionedTypedObject) ([]byte, error)
}

func NewInternalVersionedObjectType[T VersionedTypedObject](encoder TypedObjectEncoder[T], types ...string) InternalVersionedObjectType[T] {
	return InternalVersionedObjectType[T]{
		ObjectVersionedType: NewVersionedObjectType(types...),
		encoder:             encoder,
	}
}

func (o *InternalVersionedObjectType[T]) encode(obj VersionedTypedObject) ([]byte, error) {
	// cannot type parameter here, because casts of paramerized objects are not supported in GO
	return o.encoder.Encode(obj.(T), DefaultJSONEncoding)
}

func GetEncoder[T VersionedTypedObject](obj T) encoder {
	var i interface{} = obj
	if e, ok := i.(encoder); ok {
		return e
	}
	return nil
}

func MarshalObjectVersionedType[T VersionedTypedObject](obj T, toe ...TypedObjectEncoder[T]) ([]byte, error) {
	// if e, ok := obj.(encoder); ok && e.getEncoder() != nil {
	if e := GetEncoder(obj); e != nil {
		// here parameterized type casts using T would fail
		return e.encode(obj)
	}
	if e := utils.Optional(toe...); e != nil {
		return e.Encode(obj, DefaultJSONEncoding)
	}
	return nil, errors.ErrUnknown("object type", obj.GetType())
}

func KindVersion(t string) (string, string) {
	i := strings.LastIndex(t, VersionSeparator)
	if i > 0 {
		return t[:i], t[i+1:]
	}
	return t, ""
}

type VersionedType[T VersionedTypedObject] interface {
	VersionedTypedObject
	TypedObjectDecoder[T]
}

type versionedType[T VersionedTypedObject] struct {
	ObjectVersionedType
	TypedObjectDecoder[T]
}

func NewVersionedTypeByProto[T VersionedTypedObject](name string, proto T) VersionedType[T] {
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &versionedType[T]{
		ObjectVersionedType: NewVersionedObjectType(name),
		TypedObjectDecoder:  MustNewDirectDecoder(proto),
	}
}

func NewVersionedTypeByVersion[T VersionedTypedObject](name string, version FormatVersion[T]) VersionedType[T] {
	return &versionedType[T]{
		ObjectVersionedType: NewVersionedObjectType(name),
		TypedObjectDecoder:  version,
	}
}
