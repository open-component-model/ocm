// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing

import (
	"github.com/open-component-model/ocm/pkg/signing"
)

func DefaultHandlerRegistry() signing.Registry {
	return signing.DefaultRegistry()
}
