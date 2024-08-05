package vault

import (
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var (
	REALM = ocmlog.DefineSubRealm("HashiCorp Vault Access", "credentials", "vault")
	log   = ocmlog.DynamicLogger(REALM)
)
