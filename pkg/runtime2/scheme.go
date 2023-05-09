// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/modern-go/reflect2"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils"
)

// TypeGetter is the interface to be implemented for extracting a type.
type TypeGetter interface {
	// GetType returns the type of the access object.
	GetType() string
}

// TypeSetter is the interface to be implemented for extracting a type.
type TypeSetter interface {
	// SetType sets the type of an abstract element
	SetType(typ string)
}

// TypedObject defines the accessor for a typed object with additional data.
type TypedObject interface {
	TypeGetter
}

var (
	typeTypedObject = reflect.TypeOf((*TypedObject)(nil)).Elem()
	typeUnknown     = reflect.TypeOf((*Unknown)(nil)).Elem()
)

// TypedObjectDecoder is able to provide an effective typed object for some
// serilaized form. The technical deserialization is done by an Unmarshaler.
type TypedObjectDecoder[T TypedObject] interface {
	Decode(data []byte, unmarshaler Unmarshaler) (T, error)
}

// TypedObjectEncoder is able to provide a versioned representation of
// an effective TypedObject.
type TypedObjectEncoder[T TypedObject] interface {
	Encode(T, Marshaler) ([]byte, error)
}

type DirectDecoder[T TypedObject] struct {
	proto reflect.Type
}

var _ TypedObjectDecoder[TypedObject] = &DirectDecoder[TypedObject]{}

func MustNewDirectDecoder[T TypedObject](proto T) *DirectDecoder[T] {
	d, err := NewDirectDecoder[T](proto)
	if err != nil {
		panic(err)
	}
	return d
}

func NewDirectDecoder[T TypedObject](proto T) (*DirectDecoder[T], error) {
	t := MustProtoType(proto)
	if !reflect.PtrTo(t).Implements(typeTypedObject) {
		return nil, errors.Newf("object interface %T: must implement TypedObject", proto)
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.Newf("prototype %q must be a struct", t)
	}
	return &DirectDecoder[T]{
		proto: t,
	}, nil
}

func (d *DirectDecoder[T]) CreateInstance() T {
	return reflect.New(d.proto).Interface().(T)
}

func (d *DirectDecoder[T]) Decode(data []byte, unmarshaler Unmarshaler) (T, error) {
	var zero T
	inst := d.CreateInstance()
	err := unmarshaler.Unmarshal(data, inst)
	if err != nil {
		return zero, err
	}

	return inst, nil
}

func (d *DirectDecoder[T]) Encode(obj T, marshaler Marshaler) ([]byte, error) {
	return marshaler.Marshal(obj)
}

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

// KnownTypes is a set of known type names mapped to appropriate object decoders.
type KnownTypes[T TypedObject] map[string]TypedObjectDecoder[T]

// Copy provides a copy of the actually known types.
func (t KnownTypes[T]) Copy() KnownTypes[T] {
	n := KnownTypes[T]{}
	for k, v := range t {
		n[k] = v
	}
	return n
}

// TypeNames return a sorted list of known type names.
func (t KnownTypes[T]) TypeNames() []string {
	types := make([]string, 0, len(t))
	for t := range t {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// Unknown is the interface to be implemented by
// representations on an unknown, but nevertheless decoded specification
// of a typed object.
type Unknown interface {
	IsUnknown() bool
}

func IsUnknown(o TypedObject) bool {
	if o == nil {
		return true
	}
	if u, ok := o.(Unknown); ok {
		return u.IsUnknown()
	}
	return false
}

// Scheme is the interface to describe a set of object types
// that implement a dedicated interface.
// As such it knows about the desired interface of the instances
// and can validate it. Additionally, it provides an implementation
// for generic unstructured objects that can be used to decode
// any serialized from of object candidates and provide the
// effective type.
type Scheme[T TypedObject] interface {
	RegisterByDecoder(typ string, decoder TypedObjectDecoder[T]) error

	ValidateInterface(object T) error
	CreateUnstructured() T
	Convert(object TypedObject) (TypedObject, error)
	GetDecoder(otype string) TypedObjectDecoder[T]
	Decode(data []byte, unmarshaler Unmarshaler) (T, error)
	Encode(obj T, marshaler Marshaler) ([]byte, error)
	EnforceDecode(data []byte, unmarshaler Unmarshaler) (T, error)
	KnownTypes() KnownTypes[T]
	KnownTypeNames() []string
}

type SchemeBase[T TypedObject] interface {
	AddKnownTypes(scheme Scheme[T])
	Scheme[T]
}
type defaultScheme[T TypedObject] struct {
	lock           sync.RWMutex
	base           Scheme[T]
	instance       reflect.Type
	unstructured   reflect.Type
	defaultdecoder TypedObjectDecoder[T]
	acceptUnknown  bool
	types          KnownTypes[T]
}

type BaseScheme[T TypedObject] interface {
	BaseScheme() Scheme[T]
}

var _ BaseScheme[TypedObject] = (*defaultScheme[TypedObject])(nil)

func MustNewDefaultScheme[T TypedObject](protoIfce *T, protoUnstr Unstructured, acceptUnknown bool, defaultdecoder TypedObjectDecoder[T], base ...Scheme[T]) SchemeBase[T] {
	return utils.Must(NewDefaultScheme[T](protoIfce, protoUnstr, acceptUnknown, defaultdecoder, base...))
}

func NewDefaultScheme[T TypedObject](protoIfce *T, protoUnstr Unstructured, acceptUnknown bool, defaultdecoder TypedObjectDecoder[T], base ...Scheme[T]) (SchemeBase[T], error) {
	var err error

	if reflect2.IsNil(protoIfce) {
		return nil, fmt.Errorf("object interface must be given by pointer to interacted (is nil)")
	}
	it := reflect.TypeOf(protoIfce)
	if it.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("object interface %T: must be given by pointer to interacted (is not pointer)", protoIfce)
	}
	it = it.Elem()
	if it.Kind() != reflect.Interface {
		return nil, fmt.Errorf("object interface %T: must be given by pointer to interacted (does not point to interface)", protoIfce)
	}
	if !it.Implements(typeTypedObject) {
		return nil, fmt.Errorf("object interface %T: must implement TypedObject", protoIfce)
	}

	var ut reflect.Type
	if acceptUnknown {
		ut, err = ProtoType(protoUnstr)
		if err != nil {
			return nil, errors.Wrapf(err, "unstructured prototype %T", protoUnstr)
		}
		if !reflect.PtrTo(ut).Implements(typeTypedObject) {
			return nil, fmt.Errorf("unstructured type %T must implement TypedObject to be acceptale as unknown result", protoUnstr)
		}
		if !reflect.PtrTo(ut).Implements(typeUnknown) {
			return nil, fmt.Errorf("unstructured type %T must implement Unknown to be acceptable as unknown result", protoUnstr)
		}
	}

	return &defaultScheme[T]{
		base:           utils.Optional(base...),
		instance:       it,
		unstructured:   ut,
		defaultdecoder: defaultdecoder,
		types:          KnownTypes[T]{},
		acceptUnknown:  acceptUnknown,
	}, nil
}

func (d *defaultScheme[T]) BaseScheme() Scheme[T] {
	return d.base
}

func (d *defaultScheme[T]) AddKnownTypes(s Scheme[T]) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for k, v := range s.KnownTypes() {
		d.types[k] = v
	}
}

func (d *defaultScheme[T]) KnownTypes() KnownTypes[T] {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if d.base == nil {
		return d.types.Copy()
	}
	kt := d.base.KnownTypes()
	for n, t := range d.types {
		kt[n] = t
	}
	return kt
}

// KnownTypeNames return a sorted list of known type names.
func (d *defaultScheme[T]) KnownTypeNames() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	types := make([]string, 0, len(d.types))
	for t := range d.types {
		types = append(types, t)
	}
	if d.base != nil {
		types = append(types, d.base.KnownTypeNames()...)
	}
	sort.Strings(types)
	return types
}

func RegisterByType[T TypedObject](s Scheme[T], typ string, proto T) error {
	t, err := NewDirectDecoder[T](proto)
	if err != nil {
		return err
	}
	return s.RegisterByDecoder(typ, t)
}

func (d *defaultScheme[T]) RegisterByDecoder(typ string, decoder TypedObjectDecoder[T]) error {
	if decoder == nil {
		return errors.Newf("decoder must be given")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	d.types[typ] = decoder
	return nil
}

func (d *defaultScheme[T]) ValidateInterface(object T) error {
	t := reflect.TypeOf(object)
	if !t.Implements(d.instance) {
		return errors.Newf("object type %q does not implement required instance interface %q", t, d.instance)
	}
	return nil
}

func (d *defaultScheme[T]) GetDecoder(typ string) TypedObjectDecoder[T] {
	d.lock.RLock()
	defer d.lock.RUnlock()
	decoder := d.types[typ]
	if decoder == nil && d.base != nil {
		decoder = d.base.GetDecoder(typ)
	}
	return decoder
}

func (d *defaultScheme[T]) CreateUnstructured() T {
	var zero T
	if d.unstructured == nil {
		return zero
	}
	return reflect.New(d.unstructured).Interface().(T)
}

func (d *defaultScheme[T]) Encode(obj T, marshaler Marshaler) ([]byte, error) {
	if marshaler == nil {
		marshaler = DefaultYAMLEncoding
	}
	decoder := d.GetDecoder(obj.GetType())
	if encoder, ok := decoder.(TypedObjectEncoder[T]); ok {
		return encoder.Encode(obj, marshaler)
	}
	return marshaler.Marshal(obj)
}

func (d *defaultScheme[T]) Decode(data []byte, unmarshal Unmarshaler) (T, error) {
	var zero T

	var to TypedObject
	un := d.CreateUnstructured()
	if reflect2.IsNil(un) {
		to = &UnstructuredTypedObject{}
	} else {
		to = un
	}
	if unmarshal == nil {
		unmarshal = DefaultYAMLEncoding
	}
	err := unmarshal.Unmarshal(data, to)
	if err != nil {
		return zero, errors.Wrapf(err, "cannot unmarshal unstructured")
	}
	if un.GetType() == "" {
		/*
			if d.acceptUnknown {
				return un.(TypedObject), nil
			}
		*/
		return zero, errors.Newf("no type found")
	}
	decoder := d.GetDecoder(un.GetType())
	if decoder == nil {
		if d.defaultdecoder != nil {
			o, err := d.defaultdecoder.Decode(data, unmarshal)
			if err == nil {
				if !reflect2.IsNil(o) {
					return o, nil
				}
			} else if !errors.IsErrUnknownKind(err, errors.KIND_OBJECTTYPE) {
				return zero, err
			}
		}
		if d.acceptUnknown {
			return un, nil
		}
		return zero, errors.ErrUnknown(errors.KIND_OBJECTTYPE, un.GetType())
	}
	return decoder.Decode(data, unmarshal)
}

func (d *defaultScheme[T]) EnforceDecode(data []byte, unmarshal Unmarshaler) (T, error) {
	var zero T

	un := d.CreateUnstructured()
	if unmarshal == nil {
		unmarshal = DefaultYAMLEncoding.Unmarshaler
	}
	err := unmarshal.Unmarshal(data, un)
	if err != nil {
		return zero, errors.Wrapf(err, "cannot unmarshal unstructured")
	}
	if un.GetType() == "" {
		if d.acceptUnknown {
			return un, nil
		}
		return un, errors.Newf("no type found")
	}
	decoder := d.GetDecoder(un.GetType())
	if decoder == nil {
		if d.defaultdecoder != nil {
			o, err := d.defaultdecoder.Decode(data, unmarshal)
			if err == nil {
				return o, nil
			}
			if !errors.IsErrUnknownKind(err, errors.KIND_OBJECTTYPE) {
				return un, err
			}
		}
		if d.acceptUnknown {
			return un, nil
		}
		return un, errors.ErrUnknown(errors.KIND_OBJECTTYPE, un.GetType())
	}
	o, err := decoder.Decode(data, unmarshal)
	if err != nil {
		return un, err
	}
	return o, err
}

func (d *defaultScheme[T]) Convert(o TypedObject) (TypedObject, error) {
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
			object, err := d.defaultdecoder.Decode(data, DefaultJSONEncoding)
			if err == nil {
				return object, nil
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
