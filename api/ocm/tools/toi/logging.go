package toi

import (
	"ocm.software/ocm/api/utils/logging"
)

var REALM = logging.DefineSubRealm("TOI logging", "toi")

var Log = logging.DynamicLogger(REALM)
