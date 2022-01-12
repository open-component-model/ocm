// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"encoding/json"
)

// TypedObject defines the accessor for a typed object with additional data.
type TypedObject interface {
	// GetType returns the type of the access object.
	GetType() string
	// SetType sets the type of the access object.
	SetType(ttype string)
}

// KnownTypeValidationFunc defines a function that can validate types.
type KnownTypeValidationFunc func(ttype string) error

// KnownTypes defines a set of known types.
type KnownTypes interface {
	GetCodec(otype string) TypedObjectCodec
}

type SimpleKnownTypes map[string]TypedObjectCodec

func (t SimpleKnownTypes) GetCodec(otype string) TypedObjectCodec {
	return t[otype]
}

// TypedObjectCodec describes a known component type and how it is decoded and encoded
type TypedObjectCodec interface {
	TypedObjectDecoder
	TypedObjectEncoder
}

// TypedObjectCodecWrapper is a simple struct that implements the TypedObjectCodec interface
type TypedObjectCodecWrapper struct {
	TypedObjectDecoder
	TypedObjectEncoder
}

// TypedObjectDecoder defines a decoder for a typed object.
type TypedObjectDecoder interface {
	Decode(data []byte) (TypedObject, error)
}

// TypedObjectEncoder defines a encoder for a typed object.
type TypedObjectEncoder interface {
	Encode(accessor TypedObject) ([]byte, error)
}

// TypedObjectDecoderFunc is a simple function that implements the TypedObjectDecoder interface.
type TypedObjectDecoderFunc func(data []byte) error

// Decode is the Decode implementation of the TypedObjectDecoder interface.
func (e TypedObjectDecoderFunc) Decode(data []byte) error {
	return e(data)
}

// TypedObjectEncoderFunc is a simple function that implements the TypedObjectEncoder interface.
type TypedObjectEncoderFunc func(accessor TypedObject) ([]byte, error)

// Encode is the Encode implementation of the TypedObjectEncoder interface.
func (e TypedObjectEncoderFunc) Encode(accessor TypedObject) ([]byte, error) {
	return e(accessor)
}

// DefaultJSONTypedObjectCodec implements TypedObjectCodec interface with the json decoder and json encoder.
var DefaultJSONTypedObjectCodec = TypedObjectCodecWrapper{
	TypedObjectDecoder: DefaultJSONTypedObjectDecoder{},
	TypedObjectEncoder: DefaultJSONTypedObjectEncoder{},
}

// DefaultJSONTypedObjectDecoder is a simple decoder that implements the TypedObjectDecoder interface.
// It simply decodes the access using the json marshaller.
type DefaultJSONTypedObjectDecoder struct{}

var _ TypedObjectDecoder = DefaultJSONTypedObjectDecoder{}

// Decode is the Decode implementation of the TypedObjectDecoder interface.
func (e DefaultJSONTypedObjectDecoder) Decode(data []byte) (TypedObject, error) {
	var unstructured *UnstructuredTypedObject
	err := json.Unmarshal(data, unstructured)
	if err != nil {
		return nil, err
	}
	return unstructured, nil
}

// DefaultJSONTypedObjectEncoder is a simple decoder that implements the TypedObjectDecoder interface.
// It encodes the access type using the default json marshaller.
type DefaultJSONTypedObjectEncoder struct{}

var _ TypedObjectEncoder = DefaultJSONTypedObjectEncoder{}

// Encode is the Encode implementation of the TypedObjectEncoder interface.
func (e DefaultJSONTypedObjectEncoder) Encode(obj TypedObject) ([]byte, error) {
	obj.SetType(obj.GetType()) // hardcode the correct type if the type was not correctly constructed.
	return json.Marshal(obj)
}

type codec struct {
	knownTypes     KnownTypes
	defaultCodec   TypedObjectCodec
	validationFunc KnownTypeValidationFunc
}

// NewCodec creates a new typed object codec.
func NewCodec(knownTypes KnownTypes, defaultCodec TypedObjectCodec, validationFunc KnownTypeValidationFunc) TypedObjectCodec {
	if defaultCodec == nil {
		defaultCodec = DefaultJSONTypedObjectCodec
	}

	return &codec{
		defaultCodec:   defaultCodec,
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
		if err := c.validationFunc(accessType.GetType()); err != nil {
			return nil, err
		}
	}

	codec := c.knownTypes.GetCodec(accessType.GetType())
	if codec == nil {
		codec = c.defaultCodec
	}

	return codec.Decode(data)
}

// Encode marshals a typed object into a unstructured typed object.
// The given known types are used to decode the data into a specific.
// The given defaultCodec is used if no matching type is known.
// An error is returned when the type is unknown and the default codec is nil.
func (c *codec) Encode(acc TypedObject) ([]byte, error) {
	if c.validationFunc != nil {
		if err := c.validationFunc(acc.GetType()); err != nil {
			return nil, err
		}
	}

	codec := c.knownTypes.GetCodec(acc.GetType())
	if codec == nil {
		codec = c.defaultCodec
	}

	return codec.Encode(acc)
}


// JSONTypedObjectCodecBase can be used a s base object to provide default codec implementations.
type JSONTypedObjectCodecBase struct {}

func (_ JSONTypedObjectCodecBase) Encode(obj TypedObject) ([]byte, error) {
	return DefaultJSONTypedObjectCodec.Encode(obj)
}

func (_ JSONTypedObjectCodecBase) Decode(data []byte) (TypedObject, error) {
	return DefaultJSONTypedObjectCodec.Decode(data)
}

func JSONUnmarshalInto(data []byte, obj TypedObject) (TypedObject, error) {
	err:= json.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}