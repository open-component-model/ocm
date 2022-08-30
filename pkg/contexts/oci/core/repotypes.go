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

package core

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/modern-go/reflect2"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type RepositoryType interface {
	runtime.TypedObjectDecoder
	runtime.VersionedTypedObject
}

type IntermediateRepositorySpecAspect interface {
	IsIntermediate() bool
}

type RepositorySpec interface {
	runtime.VersionedTypedObject

	Name() string
	UniformRepositorySpec() *UniformRepositorySpec
	Repository(Context, credentials.Credentials) (Repository, error)
}

type RepositoryTypeScheme interface {
	runtime.Scheme

	GetRepositoryType(name string) RepositoryType
	Register(name string, atype RepositoryType)

	DecodeRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error)
	CreateRepositorySpec(obj runtime.TypedObject) (RepositorySpec, error)
}

type repositoryTypeScheme struct {
	runtime.Scheme
}

func NewRepositoryTypeScheme(defaultRepoDecoder runtime.TypedObjectDecoder) RepositoryTypeScheme {
	var rt RepositorySpec
	scheme := runtime.MustNewDefaultScheme(&rt, &UnknownRepositorySpec{}, true, defaultRepoDecoder)
	return &repositoryTypeScheme{scheme}
}

func (t *repositoryTypeScheme) AddKnowntypes(s RepositoryTypeScheme) {
	t.Scheme.AddKnownTypes(s)
}

func (t *repositoryTypeScheme) GetRepositoryType(name string) RepositoryType {
	d := t.GetDecoder(name)
	if d == nil {
		return nil
	}
	return d.(RepositoryType)
}

func (t *repositoryTypeScheme) RegisterByDecoder(name string, decoder runtime.TypedObjectDecoder) error {
	if _, ok := decoder.(RepositoryType); !ok {
		return errors.ErrInvalid("type", reflect.TypeOf(decoder).String())
	}
	return t.Scheme.RegisterByDecoder(name, decoder)
}

func (t *repositoryTypeScheme) AddKnownTypes(scheme runtime.Scheme) error {
	if _, ok := scheme.(RepositoryTypeScheme); !ok {
		return errors.ErrInvalid("type", reflect.TypeOf(scheme).String(), "expected", "RepositoryTypeScheme")
	}

	if err := t.Scheme.AddKnownTypes(scheme); err != nil {
		return fmt.Errorf("failed to add known type in repository type scheme: %w", err)
	}

	return nil
}

func (t *repositoryTypeScheme) Register(name string, rtype RepositoryType) {
	t.Scheme.RegisterByDecoder(name, rtype)
}

func (t *repositoryTypeScheme) DecodeRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	obj, err := t.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	if spec, ok := obj.(RepositorySpec); ok {
		return spec, nil
	}
	return nil, fmt.Errorf("invalid access spec type: yield %T instead of RepositorySpec", obj)
}

func (t *repositoryTypeScheme) CreateRepositorySpec(obj runtime.TypedObject) (RepositorySpec, error) {
	if s, ok := obj.(RepositorySpec); ok {
		r, err := t.Scheme.Convert(s)
		if err != nil {
			return nil, err
		}
		return r.(RepositorySpec), nil
	}
	if u, ok := obj.(*runtime.UnstructuredTypedObject); ok {
		raw, err := u.GetRaw()
		if err != nil {
			return nil, err
		}
		return t.DecodeRepositorySpec(raw, runtime.DefaultJSONEncoding)
	}
	return nil, fmt.Errorf("invalid object type %T for repository specs", obj)
}

// DefaultRepositoryTypeScheme contains all globally known access serializer.
var DefaultRepositoryTypeScheme = NewRepositoryTypeScheme(nil)

func RegisterRepositoryType(name string, atype RepositoryType) {
	DefaultRepositoryTypeScheme.Register(name, atype)
}

func CreateRepositorySpec(t runtime.TypedObject) (RepositorySpec, error) {
	return DefaultRepositoryTypeScheme.CreateRepositorySpec(t)
}

type UnknownRepositorySpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var _ RepositorySpec = &UnknownRepositorySpec{}

func (r *UnknownRepositorySpec) Name() string {
	return "unknown-" + r.GetKind()
}

func (r *UnknownRepositorySpec) UniformRepositorySpec() *UniformRepositorySpec {
	return UniformRepositorySpecForUnstructured(&r.UnstructuredVersionedTypedObject)
}

func (r *UnknownRepositorySpec) Repository(Context, credentials.Credentials) (Repository, error) {
	return nil, errors.ErrUnknown("repository type", r.GetType())
}

////////////////////////////////////////////////////////////////////////////////

type GenericRepositorySpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

func (s *GenericRepositorySpec) Name() string {
	return "generic-" + s.GetKind()
}

func (s *GenericRepositorySpec) UniformRepositorySpec() *UniformRepositorySpec {
	return UniformRepositorySpecForUnstructured(&s.UnstructuredVersionedTypedObject)
}

func (s *GenericRepositorySpec) Evaluate(ctx Context) (RepositorySpec, error) {
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	return ctx.RepositoryTypes().DecodeRepositorySpec(raw, runtime.DefaultJSONEncoding)
}

func (s *GenericRepositorySpec) Repository(ctx Context, creds credentials.Credentials) (Repository, error) {
	spec, err := s.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return spec.Repository(ctx, creds)
}

var _ RepositorySpec = &GenericRepositorySpec{}

func ToGenericRepositorySpec(spec RepositorySpec) (*GenericRepositorySpec, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	if g, ok := spec.(*GenericRepositorySpec); ok {
		return g, nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return newGenericRepositorySpec(data, runtime.DefaultJSONEncoding)
}

func NewGenericRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (RepositorySpec, error) {
	s, err := newGenericRepositorySpec(data, unmarshaler)
	if err != nil {
		return nil, err // GO is great
	}
	return s, nil
}

func newGenericRepositorySpec(data []byte, unmarshaler runtime.Unmarshaler) (*GenericRepositorySpec, error) {
	unstr := &runtime.UnstructuredVersionedTypedObject{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, unstr)
	if err != nil {
		return nil, err
	}
	return &GenericRepositorySpec{*unstr}, nil
}

////////////////////////////////////////////////////////////////////////////////
