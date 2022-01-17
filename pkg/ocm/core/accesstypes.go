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

package core

import (
	"fmt"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

type AccessType interface {
	runtime.TypedObjectCodec
	common.VersionedElement
}

type AccessSpec interface {
	compdesc.AccessSpec
	ValidFor(repo Repository) bool
	AccessMethod(access ComponentAccess) (AccessMethod, error)
}

type AccessMethod interface {
	GetName() string
	BlobAccess
}

type KnownAccessTypes interface {
	runtime.TypedObjectCodec

	GetCodec(name string) runtime.TypedObjectCodec

	GetAccessType(name string) AccessType
	Register(name string, atype AccessType)

	DecodeAccessSpec(data []byte) (AccessSpec, error)
	CreateAccessSpec(obj runtime.TypedObject) (AccessSpec, error)
}

type knownAccessTypes struct {
	runtime.TypedObjectCodec
	types map[string]AccessType
}

func NewKnownAccessTypes() KnownAccessTypes {
	types := &knownAccessTypes{
		types: map[string]AccessType{},
	}
	types.TypedObjectCodec = runtime.NewCodec(types, defaultJSONAccessSpecCodec, nil)
	return types
}

func (t *knownAccessTypes) GetCodec(name string) runtime.TypedObjectCodec {
	return t.types[name]
}

func (t *knownAccessTypes) GetAccessType(name string) AccessType {
	return t.types[name]
}

func (t *knownAccessTypes) Register(name string, atype AccessType) {
	t.types[name] = atype
}

func (t *knownAccessTypes) DecodeAccessSpec(data []byte) (AccessSpec, error) {
	obj, err := t.Decode(data)
	if err != nil {
		return nil, err
	}
	if spec, ok := obj.(AccessSpec); ok {
		return spec, nil
	}
	return nil, fmt.Errorf("invalid access spec type: yield %T instead of AccessSpec")
}

func (t *knownAccessTypes) CreateAccessSpec(obj runtime.TypedObject) (AccessSpec, error) {
	if s, ok := obj.(AccessSpec); ok {
		return s, nil
	}
	if u, ok := obj.(*runtime.UnstructuredTypedObject); ok {
		raw, err := u.GetRaw()
		if err != nil {
			return nil, err
		}
		return t.DecodeAccessSpec(raw)
	}
	return nil, errors.ErrInvalid("object type", fmt.Sprintf("%T", obj), "access specs")
}

// DefaultKnownAccessTypes contains all globally known access serializer
var DefaultKnownAccessTypes = NewKnownAccessTypes()

func RegisterAccessType(atype AccessType) {
	DefaultKnownAccessTypes.Register(atype.GetName(), atype)
}

func GetAccessType(name string) AccessType {
	return DefaultKnownAccessTypes.GetAccessType(name)
}

func CreateAccessSpec(t runtime.TypedObject) (AccessSpec, error) {
	return DefaultKnownAccessTypes.CreateAccessSpec(t)
}

// DefaultJSONTAccessSpecDecoder is a simple decoder that implements the TypedObjectDecoder interface.
// It simply decodes the access using the json marshaller.
type DefaultJSONAccessSpecDecoder struct{}

// Decode is the Decode implementation of the TypedObjectDecoder interface.
func (e DefaultJSONAccessSpecDecoder) Decode(data []byte) (runtime.TypedObject, error) {
	obj, err := runtime.DefaultJSONTypedObjectCodec.Decode(data)
	if err != nil {
		return nil, err
	}
	return &UnknownAccessSpec{&runtime.UnstructuredVersionedTypedObject{obj.(*runtime.UnstructuredTypedObject)}}, nil
}

// defaultJSONAccessSpecCodec implements TypedObjectCodec interface with the json decoder and json encoder.
var defaultJSONAccessSpecCodec = runtime.TypedObjectCodecWrapper{
	TypedObjectDecoder: DefaultJSONAccessSpecDecoder{},
	TypedObjectEncoder: runtime.DefaultJSONTypedObjectEncoder{},
}

////////////////////////////////////////////////////////////////////////////////

type UnknownAccessSpec struct {
	*runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

func (s *UnknownAccessSpec) AccessMethod(ComponentAccess) (AccessMethod, error) {
	return nil, errors.ErrUnknown(errors.KIND_ACCESSMETHOD, s.GetType())
}

func (_ *UnknownAccessSpec) ValidFor(Repository) bool {
	return false
}

var _ AccessSpec = &UnknownAccessSpec{}

////////////////////////////////////////////////////////////////////////////////

type GenericAccessSpec struct {
	*runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

func (s *GenericAccessSpec) Evaluate(ctx Context) (AccessSpec, error) {
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	return ctx.AccessMethods().DecodeAccessSpec(raw)
}

func (s *GenericAccessSpec) AccessMethod(acc ComponentAccess) (AccessMethod, error) {
	spec, err := s.Evaluate(acc.GetContext())
	if err != nil {
		return nil, err
	}
	return spec.AccessMethod(acc)
}

func (s *GenericAccessSpec) ValidFor(repo Repository) bool {
	spec, err := s.Evaluate(repo.GetContext())
	if err != nil {
		return false
	}
	return spec.ValidFor(repo)
}

var _ AccessSpec = &GenericAccessSpec{}
