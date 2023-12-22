// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/internal"
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
