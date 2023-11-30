// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
)

func Versions() accspeccpi.AccessTypeVersionScheme {
	return versions
}
