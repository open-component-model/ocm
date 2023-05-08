// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type AccessTypeVersionScheme interface {
	Register(t AccessType) error
	AddToScheme(scheme AccessTypeScheme)
	runtime.TypedObjectEncoder
	runtime.TypedObjectDecoder
}

type accessTypeVersionScheme struct {
	kind   string
	scheme AccessTypeScheme
}

func NewAccessTypeVersionScheme(kind string) AccessTypeVersionScheme {
	return &accessTypeVersionScheme{kind, internal.NewStrictAccessTypeScheme()}
}

func (s *accessTypeVersionScheme) Register(t AccessType) error {
	if t.GetKind() != s.kind {
		return errors.ErrInvalid("access spec type", t.GetType(), "kind", s.kind)
	}
	s.scheme.Register(t.GetType(), t)
	return nil
}

func (s *accessTypeVersionScheme) AddToScheme(scheme AccessTypeScheme) {
	scheme.AddKnownTypes(s.scheme)
}

func (s *accessTypeVersionScheme) Encode(obj runtime.TypedObject, m runtime.Marshaler) ([]byte, error) {
	return s.scheme.Encode(obj, m)
}

func (s *accessTypeVersionScheme) Decode(data []byte, unmarshaler runtime.Unmarshaler) (runtime.TypedObject, error) {
	return s.scheme.Decode(data, unmarshaler)
}

func RegisterAccessType(atype AccessType) {
	internal.DefaultAccessTypeScheme.Register(atype.GetType(), atype)
}

func RegisterAccessTypeVersions(s AccessTypeVersionScheme) {
	s.AddToScheme(internal.DefaultAccessTypeScheme)
}
