package transfer

import (
	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/ocm"
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("OCM transfer handling", "transfer")

type ContextProvider interface {
	GetContext() ocm.Context
}

func Logger(c ContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.GetContext().Logger(REALM).WithValues(keyValuePairs...)
}
