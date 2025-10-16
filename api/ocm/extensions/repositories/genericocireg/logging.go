package genericocireg

import (
	"github.com/mandelsoft/logging"

	ocmlog "ocm.software/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("OCM to OCI Registry Mapping", "oci", "mapping")

var TAG_CDDIFF = logging.DefineTag("cd-diff", "component descriptor modification")

func Logger(ctx logging.ContextProvider, messageContext ...logging.MessageContext) logging.Logger {
	return ctx.LoggingContext().Logger(append([]logging.MessageContext{REALM}, messageContext...))
}
