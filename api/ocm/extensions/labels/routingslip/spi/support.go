package spi

import (
	"ocm.software/ocm/api/utils/cobrautils/flagsets/flagsetscheme"
	"ocm.software/ocm/api/utils/runtime"
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

func NewEntryType[I Entry](name string, opts ...EntryTypeOption) EntryType {
	return flagsetscheme.NewVersionedTypedObjectTypeObject[Entry](runtime.NewVersionedTypedObjectType[Entry, I](name), opts...)
}

func NewEntryTypeByConverter[I Entry, V runtime.VersionedTypedObject](name string, converter runtime.Converter[I, V], opts ...EntryTypeOption) EntryType {
	return flagsetscheme.NewVersionedTypedObjectTypeObject[Entry](runtime.NewVersionedTypedObjectTypeByConverter[Entry, I, V](name, converter), opts...)
}

func NewEntryTypeByFormatVersion(name string, fmt runtime.FormatVersion[Entry], opts ...EntryTypeOption) EntryType {
	return flagsetscheme.NewVersionedTypedObjectTypeObject[Entry](runtime.NewVersionedTypedObjectTypeByFormatVersion[Entry](name, fmt), opts...)
}

////////////////////////////////////////////////////////////////////////////////

func Register(atype EntryType) {
	DefaultEntryTypeScheme().Register(atype)
}

func RegisterEntryTypeVersions(s EntryTypeVersionScheme) {
	DefaultEntryTypeScheme().AddKnownTypes(s)
}
