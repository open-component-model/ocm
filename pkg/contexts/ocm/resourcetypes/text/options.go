// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/data"
)

type (
	Option  = data.Option
	Options = data.Options
)

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return data.WithHint(h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return data.WithGlobalAccess(a)
}

////////////////////////////////////////////////////////////////////////////////
// Local Options

func WithimeType(mime string) Option {
	return data.WithMimeType(mime)
}
