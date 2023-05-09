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

type accessType struct {
	runtime.VersionedType
	description string
	format      string
	handler     flagsets.ConfigOptionTypeSetHandler
}

func NewAccessSpecType(name string, proto AccessSpec, opts ...AccessSpecTypeOption) AccessType {
	t := accessTypeTarget{&accessType{
		VersionedType: runtime.NewVersionedTypeByProto[AccessSpec](name, proto),
	}}
	for _, o := range opts {
		o.ApplyToAccessSpecOptionTarget(t)
	}
	return t.accessType
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

type (
	AccessSpecConverter = runtime.Converter[AccessSpec]
	AccessSpecVersion   = runtime.FormatVersion[AccessSpec]
)

func NewAccessSpecVersion(proto runtime.VersionedTypedObject, converter AccessSpecConverter) AccessSpecVersion {
	return runtime.NewProtoBasedVersion[AccessSpec](proto, converter)
}

////////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////////

type ConvertedAccessType struct {
	// AccessSpecVersion // does not work with Goland
	*runtime.ConvertedType[AccessSpec]
	accessType accessType
}

var (
	_ AccessSpecVersion          = &ConvertedAccessType{}
	_ runtime.TypedObjectEncoder = &ConvertedAccessType{}
	_ AccessType                 = &ConvertedAccessType{}
)

func NewConvertedAccessSpecType(name string, v AccessSpecVersion, opts ...AccessSpecTypeOption) *ConvertedAccessType {
	ct := runtime.NewConvertedType(name, v)
	t := &ConvertedAccessType{
		ConvertedType: ct,
		accessType: accessType{
			VersionedType: ct.VersionedType,
		},
	}
	for _, o := range opts {
		o.ApplyToAccessSpecOptionTarget(accessTypeTarget{&t.accessType})
	}
	return t
}

// forward additional AccessType methods

func (c *ConvertedAccessType) ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler {
	return c.accessType.ConfigOptionTypeSetHandler()
}

func (c *ConvertedAccessType) Description() string {
	return c.accessType.Description()
}

func (c *ConvertedAccessType) Format(cli bool) string {
	return c.accessType.Format(cli)
}
