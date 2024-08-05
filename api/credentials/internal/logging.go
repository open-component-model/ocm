package internal

import (
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var (
	REALM = ocmlog.DefineSubRealm("Credentials", "credentials")
	log   = ocmlog.DynamicLogger(REALM)
)
