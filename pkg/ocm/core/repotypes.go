// Copyright 2022 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
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
	"strings"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

type RepositoryType interface {
	runtime.TypedObjectCodec
	common.VersionedElement
}

type RepositorySpec interface {
	runtime.TypedObject
	common.VersionedElement

	Repository() (Repository, error)
}

type KnownRepositoryTypes interface {
	runtime.TypedObjectCodec

	GetRepositoryType(name string) RepositoryType
	Register(name string, atype RepositoryType)

	DecodeRepositorySpec(data []byte) (RepositorySpec, error)
	CreateRepositorySpec(obj runtime.TypedObject) (RepositorySpec, error)
}

type knownRepositoryTypes struct {
	runtime.TypedObjectCodec
	types map[string]RepositoryType
}

func NewKnownRepositoryTypes() KnownRepositoryTypes {
	types := &knownRepositoryTypes{
		types: map[string]RepositoryType{},
	}
	types.TypedObjectCodec = runtime.NewCodec(types, defaultJSONRepositoryCodec, nil)
	return types
}

func (t *knownRepositoryTypes) GetCodec(name string) runtime.TypedObjectCodec {
	return t.types[name]
}

func (t *knownRepositoryTypes) GetRepositoryType(name string) RepositoryType {
	return t.types[name]
}

func (t *knownRepositoryTypes) Register(name string, rtype RepositoryType) {
	t.types[name] = rtype
}

func (t *knownRepositoryTypes) DecodeRepositorySpec(data []byte) (RepositorySpec, error) {
	obj, err := t.Decode(data)
	if err != nil {
		return nil, err
	}
	if spec, ok := obj.(RepositorySpec); ok {
		return spec, nil
	}
	return nil, fmt.Errorf("invalid access spec type: yield %T instead of RepositorySpec")
}

func (t *knownRepositoryTypes) CreateRepositorySpec(obj runtime.TypedObject) (RepositorySpec, error) {
	if s, ok := obj.(RepositorySpec); ok {
		return s, nil
	}
	if u, ok := obj.(*runtime.UnstructuredTypedObject); ok {
		raw, err := u.GetRaw()
		if err != nil {
			return nil, err
		}
		return t.DecodeRepositorySpec(raw)
	}
	return nil, fmt.Errorf("invalid object type %T for repository specs", obj)
}

// DefaultKnownAccessTypes contains all globally known access serializer
var DefaultKnownRepositoryTypes = NewKnownRepositoryTypes()

func RegisterRepositoryType(name string, atype RepositoryType) {
	DefaultKnownRepositoryTypes.Register(name, atype)
}

func CreateRepositorySpec(t runtime.TypedObject) (RepositorySpec, error) {
	return DefaultKnownRepositoryTypes.CreateRepositorySpec(t)
}

// DefaultJSONTRepositorySpecDecoder is a simple decoder that implements the TypedObjectDecoder interface.
// It simply decodes the access using the json marshaller.
type DefaultJSONRepositorySpecDecoder struct{}

// Decode is the Decode implementation of the TypedObjectDecoder interface.
func (e DefaultJSONRepositorySpecDecoder) Decode(data []byte) (runtime.TypedObject, error) {
	obj, err := runtime.DefaultJSONTypedObjectCodec.Decode(data)
	if err != nil {
		return nil, err
	}
	return &UnknownRepositorySpec{obj.(*runtime.UnstructuredTypedObject)}, nil
}

// defaultJSONTypedObjectCodec implements TypedObjectCodec interface with the json decoder and json encoder.
var defaultJSONRepositoryCodec = runtime.TypedObjectCodecWrapper{
	TypedObjectDecoder: DefaultJSONRepositorySpecDecoder{},
	TypedObjectEncoder: runtime.DefaultJSONTypedObjectEncoder{},
}

type UnknownRepositorySpec struct {
	*runtime.UnstructuredTypedObject
}

var _ RepositorySpec = &UnknownRepositorySpec{}

func (r *UnknownRepositorySpec) Repository() (Repository, error) {
	return nil, fmt.Errorf("unknown respository type %q", r.GetType())
}

func (r *UnknownRepositorySpec) GetName() string {
	t := r.GetType()
	i := strings.LastIndex(t, "/")
	if i < 0 {
		return t
	}
	return t[:i]
}

func (r *UnknownRepositorySpec) GetVersion() string {
	t := r.GetType()
	i := strings.LastIndex(t, "/")
	if i < 0 {
		return "v1"
	}
	return t[i+1:]
}
