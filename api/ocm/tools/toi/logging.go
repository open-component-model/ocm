package toi

import (
	logging2 "github.com/open-component-model/ocm/api/utils/logging"
)

var REALM = logging2.DefineSubRealm("TOI logging", "toi")

var Log = logging2.DynamicLogger(REALM)
