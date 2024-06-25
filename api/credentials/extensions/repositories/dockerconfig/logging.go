package dockerconfig

import (
	ocmlog "github.com/open-component-model/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("docker config handling as credential repository", "credentials/dockerconfig")
