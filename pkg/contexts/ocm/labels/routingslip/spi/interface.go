// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package spi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type (
	Context         = cpi.Context
	Entry           = internal.Entry
	EntryType       = internal.EntryType
	EntryTypeScheme = internal.EntryTypeScheme
)

func NewStrictEntryTypeScheme() runtime.VersionedTypeRegistry[Entry, EntryType] {
	return internal.NewStrictEntryTypeScheme()
}

func DefaultEntryTypeScheme() EntryTypeScheme {
	return internal.DefaultEntryTypeScheme()
}
