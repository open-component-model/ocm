package internal

import (
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
)

var (
	REALM = ocmlog.DefineSubRealm("Credentials", "credentials")
	log   = ocmlog.DynamicLogger(REALM)
)
