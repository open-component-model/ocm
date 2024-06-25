package internal

import (
	ocmlog "github.com/open-component-model/ocm/api/utils/logging"
)

var Realm = ocmlog.DefineSubRealm("configuration management", "config")

var Logger = ocmlog.DynamicLogger(Realm)
