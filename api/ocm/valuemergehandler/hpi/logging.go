package hpi

import (
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("value merge handling", "valuemerge")

var Log = ocmlog.DynamicLogger(REALM)
