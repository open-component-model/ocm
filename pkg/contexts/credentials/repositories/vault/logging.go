package vault

import (
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
)

var (
	REALM = ocmlog.DefineSubRealm("HashiCorp Vault Access", "credentials", "vault")
	log   = ocmlog.DynamicLogger(REALM)
)
