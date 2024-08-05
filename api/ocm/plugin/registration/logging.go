package registration

import (
	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/ocm/plugin/descriptor"
)

func Logger(c logging.ContextProvider, keyValuePairs ...interface{}) logging.Logger {
	return c.LoggingContext().Logger(descriptor.REALM).WithValues(keyValuePairs...)
}
