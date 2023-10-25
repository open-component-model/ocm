// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package generic

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/rpi"
)

type (
	Options = rpi.Options
	Option  = rpi.Option
)

func WithHint(h string) Option {
	return rpi.WithHint(h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return rpi.WithGlobalAccess(a)
}
