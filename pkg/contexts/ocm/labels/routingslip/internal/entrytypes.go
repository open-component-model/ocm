// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"fmt"

	"github.com/modern-go/reflect2"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Context = cpi.Context

type EntryType interface {
	runtime.VersionedTypedObjectType[Entry]

	Description() string
	Format() string
}

// Entry is the interface access method specifications
// must fulfill. The main task is to map the specification
// to a concrete implementation of the access method for a dedicated
// component version.
type Entry interface {
	runtime.VersionedTypedObject

	Describe(ctx Context) string
}

type (
	EntryDecoder      = runtime.TypedObjectDecoder[Entry]
	EntryTypeProvider = runtime.KnownTypesProvider[Entry, EntryType]
)

type EntryTypeScheme interface {
	runtime.TypeScheme[Entry, EntryType]
}

type _EntryTypeScheme = runtime.TypeScheme[Entry, EntryType]

type entryTypeScheme struct {
	_EntryTypeScheme
}

func NewEntryTypeScheme(base ...EntryTypeScheme) EntryTypeScheme {
	scheme := runtime.MustNewDefaultTypeScheme[Entry, EntryType](&UnknownEntry{}, true, nil, utils.Optional(base...))
	return &entryTypeScheme{scheme}
}

func NewStrictEntryTypeScheme(base ...EntryTypeScheme) runtime.VersionedTypeRegistry[Entry, EntryType] {
	scheme := runtime.MustNewDefaultTypeScheme[Entry, EntryType](nil, false, nil, utils.Optional(base...))
	return &entryTypeScheme{scheme}
}

func (t *entryTypeScheme) KnownTypes() runtime.KnownTypes[Entry, EntryType] {
	return t._EntryTypeScheme.KnownTypes() // Goland
}

func CreateEntry(t runtime.TypedObject) (Entry, error) {
	return defaultEntryTypeScheme.Convert(t)
}

////////////////////////////////////////////////////////////////////////////////

type UnknownEntry struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var (
	_ runtime.TypedObject = &UnknownEntry{}
	_ runtime.Unknown     = &UnknownEntry{}
)

func (_ *UnknownEntry) IsUnknown() bool {
	return true
}

func (u *UnknownEntry) Describe(ctx Context) string {
	return fmt.Sprintf("unknown routing slip entry %q", u.GetKind())
}

var _ Entry = &UnknownEntry{}

////////////////////////////////////////////////////////////////////////////////

type EvaluatableEntry interface {
	Entry
	Evaluate(ctx Context) (Entry, error)
}

type GenericEntry struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
}

var _ Entry = &GenericEntry{}

func ToGenericEntry(spec Entry) (*GenericEntry, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	if g, ok := spec.(*GenericEntry); ok {
		return g, nil
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	return newGenericEntry(data, runtime.DefaultJSONEncoding)
}

func NewGenericEntry(data []byte, unmarshaler ...runtime.Unmarshaler) (Entry, error) {
	return generics.AsE[Entry](newGenericEntry(data, utils.Optional(unmarshaler...)))
}

func newGenericEntry(data []byte, unmarshaler runtime.Unmarshaler) (*GenericEntry, error) {
	unstr := &runtime.UnstructuredVersionedTypedObject{}
	if unmarshaler == nil {
		unmarshaler = runtime.DefaultYAMLEncoding
	}
	err := unmarshaler.Unmarshal(data, unstr)
	if err != nil {
		return nil, err
	}
	return &GenericEntry{*unstr}, nil
}

func (s *GenericEntry) Describe(ctx Context) string {
	eff, err := s.Evaluate(ctx)
	if err != nil {
		return fmt.Sprintf("invalid access specification: %s", err.Error())
	}
	return eff.Describe(ctx)
}

func (s *GenericEntry) Evaluate(ctx Context) (Entry, error) {
	raw, err := s.GetRaw()
	if err != nil {
		return nil, err
	}
	return defaultEntryTypeScheme.Decode(raw, runtime.DefaultJSONEncoding) // TODO: switch to context
}

// defaultEntryTypeScheme contains all globally known access serializer.
var defaultEntryTypeScheme = NewEntryTypeScheme()

func DefaultEntryTypeScheme() EntryTypeScheme {
	return defaultEntryTypeScheme
}
