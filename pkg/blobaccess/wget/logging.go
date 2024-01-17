package wget

import (
	"github.com/mandelsoft/logging"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
)

var REALM = ocmlog.DefineSubRealm("blob access for wget", "blobaccess/wget")

func Logger(c logging.Context, keyValuePairs ...interface{}) logging.Logger {
	if c != nil {
		return c.Logger(REALM).WithValues(keyValuePairs...)
	} else {
		return ocmlog.Logger(REALM).WithValues(keyValuePairs)
	}
}
