// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package wget

import (
	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
)

var REALM = ocmlog.DefineSubRealm("access method for wget", "accessmethod/wget")

type ContextProvider interface {
	GetContext() cpi.Context
}

func Logger(c ContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.GetContext().Logger(REALM).WithValues(keyValuePairs...)
}
