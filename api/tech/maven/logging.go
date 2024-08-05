package maven

import "ocm.software/ocm/api/utils/logging"

var REALM = logging.DefineSubRealm("Maven repository", "maven")

var Log = logging.DynamicLogger(REALM)
