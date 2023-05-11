// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
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
	return &accessTypeVersionScheme{kind, newStrictAccessTypeScheme()}
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
	defaultAccessTypeScheme.Register(atype.GetType(), atype)
}

func RegisterAccessTypeVersions(s AccessTypeVersionScheme) {
	s.AddToScheme(defaultAccessTypeScheme)
}

////////////////////////////////////////////////////////////////////////////////

type additionalTypeInfo interface {
	ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler
	Description() string
	Format(cli bool) string
}

type accessType struct {
	runtime.VersionedTypedObjectType
	description string
	format      string
	handler     flagsets.ConfigOptionTypeSetHandler
}

var _ additionalTypeInfo = (*accessType)(nil)

func newAccessSpecType(vt runtime.VersionedTypedObjectType, opts []AccessSpecTypeOption) AccessType {
	t := accessTypeTarget{&accessType{
		VersionedTypedObjectType: vt,
	}}
	for _, o := range opts {
		o.ApplyToAccessSpecOptionTarget(t)
	}
	return t.accessType
}

func NewAccessSpecType(name string, proto AccessSpec, opts ...AccessSpecTypeOption) AccessType {
	return newAccessSpecType(runtime.NewVersionedTypedObjectTypeByProto[AccessSpec](name, proto), opts)
}

func NewAccessSpecTypeByConverter(name string, proto runtime.VersionedTypedObject, converter runtime.Converter[AccessSpec], opts ...AccessSpecTypeOption) AccessType {
	return newAccessSpecType(runtime.NewVersionedTypedObjectTypeByConverter[AccessSpec](name, proto, converter), opts)
}

func NewAccessSpecTypeByVersion(name string, version runtime.FormatVersion[AccessSpec], opts ...AccessSpecTypeOption) AccessType {
	return newAccessSpecType(runtime.NewVersionedTypedObjectTypeByVersion[AccessSpec](name, version), opts)
}

func (t *accessType) ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler {
	return t.handler
}

func (t *accessType) Description() string {
	return t.description
}

func (t *accessType) Format(cli bool) string {
	group := ""
	if t.handler != nil && cli {
		opts := t.handler.OptionTypeNames()
		var names []string
		if len(opts) > 0 {
			for _, o := range opts {
				names = append(names, "<code>--"+o+"</code>")
			}
			group = "\nOptions used to configure fields: " + strings.Join(names, ", ")
		}
	}
	return t.format + group
}

////////////////////////////////////////////////////////////////////////////////

// accessTypeTarget is used as target for option functions, it provides
// setters for fields, which should nor be modifiable for a final type object.
type accessTypeTarget struct {
	*accessType
}

func (t accessTypeTarget) SetDescription(value string) {
	t.description = value
}

func (t accessTypeTarget) SetFormat(value string) {
	t.format = value
}

func (t accessTypeTarget) SetConfigHandler(value flagsets.ConfigOptionTypeSetHandler) {
	t.handler = value
}
