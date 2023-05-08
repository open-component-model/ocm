// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"reflect"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type RepositoryTypeVersionScheme interface {
	Register(t RepositoryType) error
	AddToScheme(scheme RepositoryTypeScheme)
	runtime.TypedObjectEncoder
	runtime.TypedObjectDecoder
}

type repositoryTypeVersionScheme struct {
	kind   string
	scheme RepositoryTypeScheme
}

func NewRepositoryTypeVersionScheme(kind string) RepositoryTypeVersionScheme {
	return &repositoryTypeVersionScheme{kind, internal.NewStrictRepositoryTypeScheme()}
}

func (s *repositoryTypeVersionScheme) Register(t RepositoryType) error {
	if t.GetKind() != s.kind {
		return errors.ErrInvalid("access spec type", t.GetType(), "kind", s.kind)
	}
	s.scheme.Register(t.GetType(), t)
	return nil
}

func (s *repositoryTypeVersionScheme) AddToScheme(scheme RepositoryTypeScheme) {
	scheme.AddKnownTypes(s.scheme)
}

func (s *repositoryTypeVersionScheme) Encode(obj runtime.TypedObject, m runtime.Marshaler) ([]byte, error) {
	return s.scheme.Encode(obj, m)
}

func (s *repositoryTypeVersionScheme) Decode(data []byte, unmarshaler runtime.Unmarshaler) (runtime.TypedObject, error) {
	return s.scheme.Decode(data, unmarshaler)
}

func RegisterRepositoryType(rtype RepositoryType) {
	internal.DefaultRepositoryTypeScheme.Register(rtype.GetType(), rtype)
}

func RegisterRepositoryTypeVersions(s RepositoryTypeVersionScheme) {
	s.AddToScheme(internal.DefaultRepositoryTypeScheme)
}

////////////////////////////////////////////////////////////////////////////////

type DefaultRepositoryType struct {
	runtime.ObjectVersionedType
	runtime.TypedObjectDecoder
	checker RepositoryAccessMethodChecker
}

type RepositoryAccessMethodChecker func(internal.Context, compdesc.AccessSpec) bool

func NewRepositoryType(name string, proto internal.RepositorySpec, checker RepositoryAccessMethodChecker) internal.RepositoryType {
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &DefaultRepositoryType{
		ObjectVersionedType: runtime.NewVersionedObjectType(name),
		TypedObjectDecoder:  runtime.MustNewDirectDecoder(proto),
		checker:             checker,
	}
}

func (t *DefaultRepositoryType) LocalSupportForAccessSpec(ctx internal.Context, a compdesc.AccessSpec) bool {
	if t.checker != nil {
		return t.checker(ctx, a)
	}
	return false
}
