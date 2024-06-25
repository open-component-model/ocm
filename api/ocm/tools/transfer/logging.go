package transfer

import (
	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/api/ocm"
	ocmlog "github.com/open-component-model/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("OCM transfer handling", "transfer")

type ContextProvider interface {
	GetContext() ocm.Context
}

func Logger(c ContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.GetContext().Logger(REALM).WithValues(keyValuePairs...)
}
