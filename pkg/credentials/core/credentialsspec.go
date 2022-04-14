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

	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/modern-go/reflect2"
)

// CredentialsSpec describes a dedicated credential provided by some repository
type CredentialsSpec interface {
	CredentialsSource
	GetCredentialsName() string
	GetRepositorySpec(Context) RepositorySpec
}

type DefaultCredentialsSpec struct {
	RepositorySpec  RepositorySpec
	CredentialsName string
}

func NewCredentialsSpec(name string, repospec RepositorySpec) CredentialsSpec {
	return &DefaultCredentialsSpec{
		RepositorySpec:  repospec,
		CredentialsName: name,
	}
}

func (s *DefaultCredentialsSpec) GetCredentialsName() string {
	return s.CredentialsName
}

func (s *DefaultCredentialsSpec) GetRepositorySpec(Context) RepositorySpec {
	return s.RepositorySpec
}

func (s *DefaultCredentialsSpec) Credentials(ctx Context, creds ...CredentialsSource) (Credentials, error) {
	return ctx.CredentialsForSpec(s, creds...)
}

// MarshalJSON implements a custom json unmarshal method
func (s DefaultCredentialsSpec) MarshalJSON() ([]byte, error) {
	ocispec, err := runtime.ToUnstructuredTypedObject(s.RepositorySpec)
	if err != nil {
		return nil, err
	}
	specdata, err := runtime.ToUnstructuredObject(struct {
		Name string `json:"credentialsName,omitempty"`
	}{Name: s.CredentialsName})

	if err != nil {
		return nil, err
	}
	return json.Marshal(specdata.FlatMerge(ocispec.Object))
}

// UnmarshalJSON implements a custom default json unmarshal method.
// It should not be used because it always used the default context.
func (s *DefaultCredentialsSpec) UnmarshalJSON(data []byte) error {
	repo, err := DefaultContext.RepositoryTypes().Decode(data, runtime.DefaultJSONEncoding)
	if err != nil {
		return err
	}

	specdata := &struct {
		Name string `json:"credentialsName,omitempty"`
	}{}
	err = json.Unmarshal(data, specdata)
	if err != nil {
		return err
	}

	s.RepositorySpec = repo.(RepositorySpec)
	s.CredentialsName = specdata.Name
	return nil
}

type GenericCredentialsSpec struct {
	RepositorySpec  *GenericRepositorySpec
	CredentialsName string
}

func ToGenericCredentialsSpec(spec CredentialsSpec) (*GenericCredentialsSpec, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	if g, ok := spec.(*GenericCredentialsSpec); ok {
		return g, nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return newGenericCredentialsSpec(data, runtime.DefaultJSONEncoding)
}

func newGenericCredentialsSpec(data []byte, unmarshaler runtime.Unmarshaler) (*GenericCredentialsSpec, error) {
	gen := &GenericCredentialsSpec{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, gen)
	if err != nil {
		return nil, err
	}
	return gen, nil
}

func NewGenericCredentialsSpec(name string, repospec *GenericRepositorySpec) *GenericCredentialsSpec {
	return &GenericCredentialsSpec{
		RepositorySpec:  repospec,
		CredentialsName: name,
	}
}

var _ CredentialsSpec = &GenericCredentialsSpec{}

func (s *GenericCredentialsSpec) GetCredentialsName() string {
	return s.CredentialsName
}

func (s *GenericCredentialsSpec) GetRepositorySpec(context Context) RepositorySpec {
	return s.RepositorySpec
}

func (s *GenericCredentialsSpec) Credentials(ctx Context, creds ...CredentialsSource) (Credentials, error) {
	return ctx.CredentialsForSpec(s, creds...)
}

// MarshalJSON implements a custom json unmarshal method
func (s GenericCredentialsSpec) MarshalJSON() ([]byte, error) {
	specdata, err := runtime.ToUnstructuredObject(struct {
		Name string `json:"credentialsName,omitempty"`
	}{Name: s.CredentialsName})

	if err != nil {
		return nil, err
	}
	return json.Marshal(specdata.FlatMerge(s.RepositorySpec.Object))
}

// UnmarshalJSON implements a custom json unmarshal method for a unstructured typed object.
func (s *GenericCredentialsSpec) UnmarshalJSON(data []byte) error {
	spec := &GenericRepositorySpec{}

	err := json.Unmarshal(data, spec)
	if err != nil {
		return err
	}
	s.CredentialsName = ""
	if name, ok := spec.Object["credentialsName"]; ok {
		if n, ok := name.(string); ok {
			s.CredentialsName = n
		}
	}

	delete(spec.Object, "credentialName")
	s.RepositorySpec = spec
	return nil
}
