package wget

import (
	"github.com/mandelsoft/logging"
	"ocm.software/ocm/api/ocm/cpi"
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("access method for wget", "accessmethod/wget")

type ContextProvider interface {
	GetContext() cpi.Context
}

func Logger(c ContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.GetContext().Logger(REALM).WithValues(keyValuePairs...)
}
