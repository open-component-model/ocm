// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package spi

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type EntryTypeVersionScheme = runtime.TypeVersionScheme[Entry, EntryType]

func NewEntryTypeVersionScheme(kind string) EntryTypeVersionScheme {
	return runtime.NewTypeVersionScheme[Entry, EntryType](kind, NewStrictEntryTypeScheme())
}

////////////////////////////////////////////////////////////////////////////////

type EntryFormatVersionRegistry = runtime.FormatVersionRegistry[Entry]

func NewEntryFormatVersionRegistry() EntryFormatVersionRegistry {
	return runtime.NewFormatVersionRegistry[Entry]()
}

func MustNewEntryMultiFormatVersion(kind string, formats EntryFormatVersionRegistry) runtime.FormatVersion[Entry] {
	return runtime.MustNewMultiFormatVersion[Entry](kind, formats)
}

////////////////////////////////////////////////////////////////////////////////

type additionalTypeInfo interface {
	ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler
	Description() string
	Format() string
}

type entryType struct {
	runtime.VersionedTypedObjectType[Entry]
	description string
	format      string
	handler     flagsets.ConfigOptionTypeSetHandler
	validator   func(Entry) error
}

var _ additionalTypeInfo = (*entryType)(nil)

func NewEntryTypeByBaseType(vt runtime.VersionedTypedObjectType[Entry], opts ...EntryTypeOption) EntryType {
	t := entryTypeTarget{&entryType{
		VersionedTypedObjectType: vt,
	}}
	for _, o := range opts {
		o.ApplyToEntryOptionTarget(t)
	}
	return t.entryType
}

func NewEntryType[I Entry](name string, opts ...EntryTypeOption) EntryType {
	return NewEntryTypeByBaseType(runtime.NewVersionedTypedObjectType[Entry, I](name), opts...)
}

func NewAccessSpecTypeByConverter[I Entry, V runtime.VersionedTypedObject](name string, converter runtime.Converter[I, V], opts ...EntryTypeOption) EntryType {
	return NewEntryTypeByBaseType(runtime.NewVersionedTypedObjectTypeByConverter[Entry, I, V](name, converter), opts...)
}

func NewEntryTypeByFormatVersion(name string, fmt runtime.FormatVersion[Entry], opts ...EntryTypeOption) EntryType {
	return NewEntryTypeByBaseType(runtime.NewVersionedTypedObjectTypeByFormatVersion[Entry](name, fmt), opts...)
}

func (t *entryType) ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler {
	return t.handler
}

func (t *entryType) Description() string {
	return t.description
}

func (t *entryType) Format() string {
	return t.format
}

func (t *entryType) Validate(e Entry) error {
	if t.validator == nil {
		return nil
	}
	return t.validator(e)
}

////////////////////////////////////////////////////////////////////////////////

// entryTypeTarget is used as target for option functions, it provides
// setters for fields, which should nor be modifiable for a final type object.
type entryTypeTarget struct {
	*entryType
}

func (t entryTypeTarget) SetDescription(value string) {
	t.description = value
}

func (t entryTypeTarget) SetFormat(value string) {
	t.format = value
}

func (t entryTypeTarget) SetConfigHandler(value flagsets.ConfigOptionTypeSetHandler) {
	t.handler = value
}

////////////////////////////////////////////////////////////////////////////////

func Register(atype EntryType) {
	DefaultEntryTypeScheme().Register(atype)
}

func RegisterEntryTypeVersions(s EntryTypeVersionScheme) {
	DefaultEntryTypeScheme().AddKnownTypes(s)
}
