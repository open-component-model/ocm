// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package clictx

import (
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx/internal"
)

type (
	Context = internal.Context
	OCI     = internal.OCI
	OCM     = internal.OCM
)

func DefaultContext() Context {
	return internal.DefaultContext
}
