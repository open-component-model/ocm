package internal

import (
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
)

var Realm = ocmlog.DefineSubRealm("configuration management", "config")

var Logger = ocmlog.DynamicLogger(Realm)
