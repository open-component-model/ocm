// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package runtime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/open-component-model/ocm/pkg/errors"
)

// TypeGetter is the interface to be implemented for extracting a type
type TypeGetter interface {
	// GetType returns the type of the access object.
	GetType() string
}

// TypeSetter is the interface to be implemented for extracting a type
type TypeSetter interface {
	// SetType sets the type of an abstract element
	SetType(typ string)
}

// TypedObject defines the accessor for a typed object with additional data.
type TypedObject interface {
	TypeGetter
}

var typeTypedObject = reflect.TypeOf((*TypedObject)(nil)).Elem()

// TypedObjectDecoder is able to provide an effective typed object for some
// serilaized form. The technical deserialization is done by an Unmarshaler.
type TypedObjectDecoder interface {
	Decode(data []byte, unmarshaler Unmarshaler) (TypedObject, error)
}

// TypedObjectEncoder is able to provide a versioned representation of
// of an effective TypedObject
type TypedObjectEncoder interface {
	Encode(TypedObject, Marshaler) ([]byte, error)
}

type DirectDecoder struct {
	proto reflect.Type
}

var _ TypedObjectDecoder = &DirectDecoder{}

func MustNewDirectDecoder(proto interface{}) *DirectDecoder {
	d, err := NewDirectDecoder(proto)
	if err != nil {
		panic(err)
	}
	return d
}

func NewDirectDecoder(proto interface{}) (*DirectDecoder, error) {
	t := MustProtoType(proto)
	if !reflect.PtrTo(t).Implements(typeTypedObject) {
		return nil, errors.Newf("object interface %T: must implement TypedObject", proto)
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.Newf("prototype %q must be a struct", t)
	}
	return &DirectDecoder{
		proto: t,
	}, nil
}

func (d *DirectDecoder) CreateInstance() TypedObject {
	return reflect.New(d.proto).Interface().(TypedObject)
}

func (d *DirectDecoder) Decode(data []byte, unmarshaler Unmarshaler) (TypedObject, error) {
	inst := d.CreateInstance()
	err := unmarshaler.Unmarshal(data, inst)
	if err != nil {
		return nil, err
	}

	return inst.(TypedObject), nil
}

func (d *DirectDecoder) Encode(obj TypedObject, marshaler Marshaler) ([]byte, error) {
	return marshaler.Marshal(obj)
}

// TypedObjectConverter converts a versioned representation into the
// intended type required by the scheme.
type TypedObjectConverter interface {
	ConvertTo(in interface{}) (TypedObject, error)
}

// ConvertingDecoder uses a serialization from different from the
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

// KnownTypes is a set of known type names mapped to appropriate object decoders.
type KnownTypes map[string]TypedObjectDecoder

// Copy provides a copy of the actually known types
func (t KnownTypes) Copy() KnownTypes {
	n := KnownTypes{}
	for k, v := range t {
		n[k] = v
	}
	return n
}

// TypeNames return a sorted list of known type names
func (t KnownTypes) TypeNames() []string {
	types := make([]string, 0, len(t))
	for t := range t {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// Scheme is the interface to describe a set of object types
// that implement a dedicated interface.
// As such it knows about the desired interface of the instances
// and can validate it. Additionally it provides an implementation
// for generic unstructured objects that can be used to decode
// any serialized from of object candidates and provide the
// effective type.
type Scheme interface {
	RegisterByDecoder(typ string, decoder TypedObjectDecoder) error

	ValidateInterface(object TypedObject) error
	CreateUnstructured() Unstructured
	Convert(object TypedObject) (TypedObject, error)
	GetDecoder(otype string) TypedObjectDecoder
	Decode(data []byte, unmarshaler Unmarshaler) (TypedObject, error)
	Encode(obj TypedObject, marshaler Marshaler) ([]byte, error)
	EnforceDecode(data []byte, unmarshaler Unmarshaler) (TypedObject, error)
	AddKnownTypes(scheme Scheme)
	KnownTypes() KnownTypes
	KnownTypeNames() []string
}

type defaultScheme struct {
	lock           sync.RWMutex
	instance       reflect.Type
	unstructured   reflect.Type
	defaultdecoder TypedObjectDecoder
	acceptUnknown  bool
	types          KnownTypes
}

func MustNewDefaultScheme(proto_ifce interface{}, proto_unstr Unstructured, acceptUnknown bool, defaultdecoder TypedObjectDecoder) Scheme {
	s, err := NewDefaultScheme(proto_ifce, proto_unstr, acceptUnknown, defaultdecoder)
	if err != nil {
		panic(err)
	}
	return s
}

func NewDefaultScheme(proto_ifce interface{}, proto_unstr Unstructured, acceptUnknown bool, defaultdecoder TypedObjectDecoder) (Scheme, error) {
	if proto_ifce == nil {
		return nil, fmt.Errorf("object interface must be given by pointer to interace (is nil)")
	}
	it := reflect.TypeOf(proto_ifce)
	if it.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("object interface %T: must be given by pointer to interace (is not pointer)", proto_ifce)
	}
	it = it.Elem()
	if it.Kind() != reflect.Interface {
		return nil, fmt.Errorf("object interface %T: must be given by pointer to interace (does not point to interface)", proto_ifce)
	}
	if !it.Implements(typeTypedObject) {
		return nil, fmt.Errorf("object interface %T: must implement TypedObject", proto_ifce)
	}

	ut, err := ProtoType(proto_unstr)
	if err != nil {
		return nil, errors.Wrapf(err, "unstructured prototype %T", proto_unstr)
	}
	if acceptUnknown {
		if !reflect.PtrTo(ut).Implements(typeTypedObject) {
			return nil, fmt.Errorf("unstructured type %T must implement TypedObject to be acceptale as unknown result", proto_unstr)
		}
	}

	return &defaultScheme{
		instance:       it,
		unstructured:   ut,
		defaultdecoder: defaultdecoder,
		types:          KnownTypes{},
		acceptUnknown:  acceptUnknown,
	}, nil
}

func (d *defaultScheme) AddKnownTypes(s Scheme) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for k, v := range s.KnownTypes() {
		d.types[k] = v
	}
}

func (d *defaultScheme) KnownTypes() KnownTypes {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.types.Copy()
}

// KnownTypeNames return a sorted list of known type names
func (d *defaultScheme) KnownTypeNames() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	types := make([]string, 0, len(d.types))
	for t := range d.types {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func RegisterByType(s Scheme, typ string, proto TypedObject) error {
	t, err := NewDirectDecoder(proto)
	if err != nil {
		return err
	}
	return s.RegisterByDecoder(typ, t)
}

func (d *defaultScheme) RegisterByDecoder(typ string, decoder TypedObjectDecoder) error {
	if decoder == nil {
		return errors.Newf("decoder must be given")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	d.types[typ] = decoder
	return nil
}

func (d *defaultScheme) ValidateInterface(object TypedObject) error {
	t := reflect.TypeOf(object)
	if !t.Implements(d.instance) {
		return errors.Newf("object type %q does not implement required instance interface %q", t, d.instance)
	}
	return nil
}

func (d *defaultScheme) GetDecoder(typ string) TypedObjectDecoder {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.types[typ]
}

func (d *defaultScheme) CreateUnstructured() Unstructured {
	return reflect.New(d.unstructured).Interface().(Unstructured)
}

func (d *defaultScheme) Encode(obj TypedObject, marshaler Marshaler) ([]byte, error) {
	if marshaler == nil {
		marshaler = DefaultYAMLEncoding
	}
	decoder := d.GetDecoder(obj.GetType())
	if encoder, ok := decoder.(TypedObjectEncoder); ok {
		return encoder.Encode(obj, marshaler)
	}
	return marshaler.Marshal(obj)
}

func (d *defaultScheme) Decode(data []byte, unmarshal Unmarshaler) (TypedObject, error) {
	un := d.CreateUnstructured()
	if unmarshal == nil {
		unmarshal = DefaultYAMLEncoding
	}
	err := unmarshal.Unmarshal(data, un)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal unstructured")
	}
	if un.GetType() == "" {
		if d.acceptUnknown {
			return un.(TypedObject), nil
		}
		return nil, errors.Newf("no type found")
	}
	decoder := d.GetDecoder(un.GetType())
	if decoder == nil {
		if d.defaultdecoder != nil {
			o, err := d.defaultdecoder.Decode(data, unmarshal)
			if err == nil {
				return o, nil
			}
			if !errors.IsErrUnknownKind(err, errors.KIND_OBJECTTYPE) {
				return nil, err
			}
		}
		if d.acceptUnknown {
			return un.(TypedObject), nil
		}
		return nil, errors.ErrUnknown(errors.KIND_OBJECTTYPE, un.GetType())
	}
	return decoder.Decode(data, unmarshal)
}

func (d *defaultScheme) EnforceDecode(data []byte, unmarshal Unmarshaler) (TypedObject, error) {
	un := d.CreateUnstructured()
	if unmarshal == nil {
		unmarshal = DefaultYAMLEncoding.Unmarshaler
	}
	err := unmarshal.Unmarshal(data, un)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal unstructured")
	}
	if un.GetType() == "" {
		if d.acceptUnknown {
			return un.(TypedObject), nil
		}
		return un.(TypedObject), errors.Newf("no type found")
	}
	decoder := d.GetDecoder(un.GetType())
	if decoder == nil {
		if d.defaultdecoder != nil {
			o, err := d.defaultdecoder.Decode(data, unmarshal)
			if err == nil {
				return o, nil
			}
			if !errors.IsErrUnknownKind(err, errors.KIND_OBJECTTYPE) {
				return un.(TypedObject), err
			}
		}
		if d.acceptUnknown {
			return un.(TypedObject), nil
		}
		return un.(TypedObject), errors.ErrUnknown(errors.KIND_OBJECTTYPE, un.GetType())
	}
	o, err := decoder.Decode(data, unmarshal)
	if err != nil {
		return un.(TypedObject), err
	}
	return o, err
}

func (d *defaultScheme) Convert(o TypedObject) (TypedObject, error) {
	if o.GetType() == "" {
		return nil, errors.Newf("no type found")
	}
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	decoder := d.GetDecoder(o.GetType())
	if decoder == nil {
		if d.defaultdecoder != nil {
			o, err := d.defaultdecoder.Decode(data, DefaultJSONEncoding)
			if err == nil {
				return o, nil
			}
			if !errors.IsErrUnknownKind(err, errors.KIND_OBJECTTYPE) {
				return nil, err
			}
		}
		return nil, errors.ErrUnknown(errors.KIND_OBJECTTYPE, o.GetType())
	}
	r, err := decoder.Decode(data, DefaultJSONEncoding)
	if err != nil {
		return nil, err
	}
	if reflect.TypeOf(r) == reflect.TypeOf(o) {
		return o, nil
	}
	return r, nil
}

////////////////////////////////////////////////////////////////////////////////

/*
// KnownTypeValidationFunc defines a function that can validate types.
type KnownTypeValidationFunc func(ttype string) error

// KnownTypes defines a set of known types.
type KnownTypes interface {
	GetDecoder(otype string) TypedObjectDecoder
}

type SimpleKnownTypes map[string]TypedObjectDecoder

func (t SimpleKnownTypes) GetDecoder(otype string) TypedObjectDecoder {
	return t[otype]
}

// TypedObjectDecoder describes a known component type and how it is decoded and encoded
type TypedObjectDecoder interface {
	Decode(data []byte) (TypedObject, error)
}

// TypedObjectDecoderWrapper is a simple struct that implements the TypedObjectDecoder interface
type TypedObjectDecoderWrapper struct {
	TypedObjectDecoder
}

// TypedObjectDecoderFunc is a simple function that implements the XXX interface.
type TypedObjectDecoderFunc func(data []byte) error

// Decode is the Decode implementation of the XXX interface.
func (e TypedObjectDecoderFunc) Decode(data []byte) error {
	return e(data)
}

// DefaultJSONTypedObjectDecoder is a simple decoder that implements the XXX interface.
// It simply decodes the access using the json marshaller.
type DefaultJSONTypedObjectDecoder struct{}

var _ TypedObjectDecoder = DefaultJSONTypedObjectDecoder{}

// Decode is the Decode implementation of the XXX interface.
func (e DefaultJSONTypedObjectDecoder) Decode(data []byte) (TypedObject, error) {
	var unstructured *UnstructuredTypedObject
	err := json.Unmarshal(data, unstructured)
	if err != nil {
		return nil, err
	}
	return unstructured, nil
}

type codec struct {
	knownTypes     KnownTypes
	defaultCodec   TypedObjectDecoder
	validationFunc KnownTypeValidationFunc
}

// NewCodec creates a new typed object codec.
func NewCodec(knownTypes KnownTypes, defaultDecoder TypedObjectDecoder, validationFunc KnownTypeValidationFunc) TypedObjectDecoder {
	if defaultDecoder == nil {
		defaultDecoder = DefaultJSONTypedObjectDecoder{}
	}

	return &codec{
		defaultCodec:   defaultDecoder,
		knownTypes:     knownTypes,
		validationFunc: validationFunc,
	}
}

// Decode unmarshals a unstructured typed object into a TypedObject.
// The given known types are used to decode the data into a specific.
// Unknown types are decoded into UnstructuredTypesObjects.
// An error is returned when the type is unknown and the default codec is nil.
func (c *codec) Decode(data []byte) (TypedObject, error) {
	accessType := &ObjectType{}
	if err := json.Unmarshal(data, accessType); err != nil {
		return nil, err
	}

	if c.validationFunc != nil {
		if err := c.validationFunc(accessType.Algorithm()); err != nil {
			return nil, err
		}
	}

	codec := c.knownTypes.GetDecoder(accessType.Algorithm())
	if codec == nil {
		codec = c.defaultCodec
	}

	return codec.Decode(data)
}

func UnmarshalInto(data []byte, obj TypedObject) (TypedObject, error) {
	err := json.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
*/
