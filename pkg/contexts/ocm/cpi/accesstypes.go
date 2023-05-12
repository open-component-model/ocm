// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type AccessTypeVersionScheme = runtime.TypeVersionScheme[AccessSpec, AccessType]

func NewAccessTypeVersionScheme(kind string) AccessTypeVersionScheme {
	return runtime.NewTypeVersionScheme[AccessSpec, AccessType](kind, newStrictAccessTypeScheme())
}

func RegisterAccessType(atype AccessType) {
	defaultAccessTypeScheme.Register(atype)
}

func RegisterAccessTypeVersions(s AccessTypeVersionScheme) {
	defaultAccessTypeScheme.AddKnownTypes(s)
}

////////////////////////////////////////////////////////////////////////////////

type additionalTypeInfo interface {
	ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler
	Description() string
	Format(cli bool) string
}

type accessType struct {
	runtime.VersionedTypedObjectType[AccessSpec]
	description string
	format      string
	handler     flagsets.ConfigOptionTypeSetHandler
}

var _ additionalTypeInfo = (*accessType)(nil)

func newAccessSpecType(vt runtime.VersionedTypedObjectType[AccessSpec], opts []AccessSpecTypeOption) AccessType {
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

func NewAccessSpecTypeByConverter(name string, proto runtime.VersionedTypedObject, converter runtime.Converter[AccessSpec, runtime.TypedObject], opts ...AccessSpecTypeOption) AccessType {
	return newAccessSpecType(runtime.NewVersionedTypedObjectTypeByProtoConverter[AccessSpec](name, proto, converter), opts)
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
