package internal

import (
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var Realm = ocmlog.DefineSubRealm("configuration management", "config")

var Logger = ocmlog.DynamicLogger(Realm)
