package routingslip

import (
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/internal"
)

type (
	Context         = internal.Context
	ContextProvider = ocm.ContextProvider
	EntryTypeScheme = internal.EntryTypeScheme
	Entry           = internal.Entry
	GenericEntry    = internal.GenericEntry
)

type SlipAccess interface {
	Get(name string) (*RoutingSlip, error)
}

func DefaultEntryTypeScheme() EntryTypeScheme {
	return internal.DefaultEntryTypeScheme()
}

func For(ctx ContextProvider) EntryTypeScheme {
	return internal.For(ctx)
}
