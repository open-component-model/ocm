package spi

import (
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/internal"
	"ocm.software/ocm/api/utils/runtime"
)

type (
	Context         = cpi.Context
	Entry           = internal.Entry
	UnknownEntry    = internal.UnknownEntry
	GenericEntry    = internal.GenericEntry
	EntryType       = internal.EntryType
	EntryTypeScheme = internal.EntryTypeScheme
)

func NewStrictEntryTypeScheme() runtime.VersionedTypeRegistry[Entry, EntryType] {
	return internal.NewStrictEntryTypeScheme()
}

func DefaultEntryTypeScheme() EntryTypeScheme {
	return internal.DefaultEntryTypeScheme()
}

func For(ctx cpi.ContextProvider) EntryTypeScheme {
	return internal.For(ctx)
}
