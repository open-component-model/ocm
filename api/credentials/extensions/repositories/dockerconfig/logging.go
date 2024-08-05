package dockerconfig

import (
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("docker config handling as credential repository", "credentials/dockerconfig")
