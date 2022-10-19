// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
)

type OCISpecFunction = core.OCISpecFunction

func RegisterOCIImplementation(impl OCISpecFunction) {
	core.RegisterOCIImplementation(impl)
}
