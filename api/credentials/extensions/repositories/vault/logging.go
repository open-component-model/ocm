package vault

import (
	ocmlog "github.com/open-component-model/ocm/api/utils/logging"
)

var (
	REALM = ocmlog.DefineSubRealm("HashiCorp Vault Access", "credentials", "vault")
	log   = ocmlog.DynamicLogger(REALM)
)
