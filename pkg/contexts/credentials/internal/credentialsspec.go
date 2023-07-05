// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"fmt"

	"github.com/modern-go/reflect2"

	"github.com/open-component-model/ocm/pkg/runtime"
)

// CredentialsSpec describes a dedicated credential provided by some repository.
type CredentialsSpec interface {
	CredentialsSource
	GetCredentialsName() string
	GetRepositorySpec(Context) (RepositorySpec, error)
}

type DefaultCredentialsSpec struct {
	RepositorySpec  RepositorySpec
	CredentialsName string
}

const CREDNAME_ATTRIBUTE = "credentialsName"

func NewCredentialsSpec(name string, repospec RepositorySpec) CredentialsSpec {
	return &DefaultCredentialsSpec{
		RepositorySpec:  repospec,
		CredentialsName: name,
	}
}

func (s *DefaultCredentialsSpec) GetCredentialsName() string {
	return s.CredentialsName
}

func (s *DefaultCredentialsSpec) GetRepositorySpec(ctx Context) (RepositorySpec, error) {
	return ctx.RepositorySpecForSpec(s.RepositorySpec)
}

func (s *DefaultCredentialsSpec) Credentials(ctx Context, creds ...CredentialsSource) (Credentials, error) {
	return ctx.CredentialsForSpec(s, creds...)
}

// MarshalJSON implements a custom json unmarshal method.
func (s DefaultCredentialsSpec) MarshalJSON() ([]byte, error) {
	specdata, err := runtime.ToUnstructuredTypedObject(s.RepositorySpec)
	if err != nil {
		return nil, err
	}
	specdata.Object[CREDNAME_ATTRIBUTE] = s.CredentialsName
	return json.Marshal(specdata)
}

// UnmarshalJSON implements a custom default json unmarshal method.
// It should not be used because it always used the default context.
func (s *DefaultCredentialsSpec) UnmarshalJSON(data []byte) error {
	spec := &GenericRepositorySpec{}
	err := json.Unmarshal(data, spec)
	if err != nil {
		return err
	}

	// check for additional property of a credentials spec.
	s.CredentialsName = ""
	if name, ok := spec.Object[CREDNAME_ATTRIBUTE]; ok {
		if n, ok := name.(string); ok {
			s.CredentialsName = n
		}
	}
	// remove additional property from generic repository spec.
	delete(spec.Object, CREDNAME_ATTRIBUTE)

	s.RepositorySpec = spec
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

func (s *GenericCredentialsSpec) GetRepositorySpec(ctx Context) (RepositorySpec, error) {
	return ctx.RepositorySpecForSpec(s.RepositorySpec)
}

func (s *GenericCredentialsSpec) Credentials(ctx Context, creds ...CredentialsSource) (Credentials, error) {
	return ctx.CredentialsForSpec(s, creds...)
}

// MarshalJSON implements a custom json unmarshal method.
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
		return fmt.Errorf("failed to unmarshal spec data: %w", err)
	}

	s.CredentialsName = ""
	if name, ok := spec.Object["credentialsName"]; ok {
		if n, ok := name.(string); ok {
			s.CredentialsName = n
		}
	}
	delete(spec.Object, "credentialsName")

	s.RepositorySpec = spec
	return nil
}
