// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package clictx

import (
	"github.com/open-component-model/ocm/pkg/contexts/clictx/core"
)

type (
	Context = core.Context
	OCI     = core.OCI
	OCM     = core.OCM
)

func DefaultContext() Context {
	return core.DefaultContext
}
